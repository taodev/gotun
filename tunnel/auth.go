package tunnel

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"time"

	"github.com/bytedance/gopkg/lang/mcache"
)

func auth(conn net.Conn, password string) (ts int64, iv []byte, err error) {
	buf := mcache.Malloc(8 + aes.BlockSize + 32)
	defer mcache.Free(buf)

	// 写入时间戳
	ts = time.Now().Unix()
	binary.BigEndian.PutUint64(buf[:8], uint64(ts))

	// 写入随机数
	iv = buf[8 : 8+aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	// 写入sha256签名
	sumBytes := mcache.Malloc(8 + aes.BlockSize + len(password))
	defer mcache.Free(sumBytes)
	copy(sumBytes, buf[:8+aes.BlockSize])
	copy(sumBytes[8+aes.BlockSize:], password)
	sum := sha256.Sum256(sumBytes)
	copy(buf[8+aes.BlockSize:], sum[:])

	if _, err = conn.Write(buf); err != nil {
		return
	}

	return
}

func authVerify(conn net.Conn, password string) (ts int64, iv []byte, err error) {
	buf := mcache.Malloc(8 + aes.BlockSize + 32)
	defer mcache.Free(buf)

	if _, err = io.ReadFull(conn, buf); err != nil {
		return
	}

	ts = int64(binary.BigEndian.Uint64(buf[:8]))
	iv = buf[8 : 8+aes.BlockSize]
	sum := buf[8+aes.BlockSize:]

	now := time.Now().Unix()

	// 判断时间戳是否过期
	if math.Abs(float64(now-ts)) > 30 {
		err = fmt.Errorf("conn: [%v] timestamp expired", conn.RemoteAddr())
		return
	}

	// 判断sha257签名
	sumBytes := mcache.Malloc(8 + aes.BlockSize + len(password))
	defer mcache.Free(sumBytes)

	copy(sumBytes, buf[:8+aes.BlockSize])
	copy(sumBytes[8+aes.BlockSize:], password)
	newSum := sha256.Sum256(sumBytes)
	if !bytes.Equal(sum, newSum[:]) {
		err = fmt.Errorf("conn: [%v] auth failed", conn.RemoteAddr())
		return
	}

	return
}

func passwordToKey(password string, ts int64) (key []byte) {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d%s", ts, password)))
	key = sum[:]
	return
}

func newAESReader(conn net.Conn, ts int64, iv []byte) io.Reader {
	key := passwordToKey("password", ts)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}

	return cipher.StreamReader{S: cipher.NewCFBDecrypter(block, iv), R: conn}
}

func newAESWriter(conn net.Conn, ts int64, iv []byte) io.Writer {
	key := passwordToKey("password", ts)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}

	return cipher.StreamWriter{S: cipher.NewCFBEncrypter(block, iv), W: conn}
}
