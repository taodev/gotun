package tunnel

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/golang/snappy"
)

const DEFAULT_TIMEOUT = 60 * time.Second

type ConnAES struct {
	net.Conn
	Reader       io.Reader
	Writer       io.Writer
	snappyReader *snappy.Reader
	snappyWriter *snappy.Writer
}

func (c *ConnAES) Read(b []byte) (n int, err error) {
	c.SetReadDeadline(time.Now().Add(DEFAULT_TIMEOUT))
	if c.snappyReader != nil {
		return c.snappyReader.Read(b)
	}

	return c.Reader.Read(b)
}

func (c *ConnAES) Write(b []byte) (n int, err error) {
	c.SetWriteDeadline(time.Now().Add(DEFAULT_TIMEOUT))
	if c.snappyWriter != nil {
		n, err = c.snappyWriter.Write(b)
		if err != nil {
			return
		}

		if err = c.snappyWriter.Flush(); err != nil {
			return
		}

		return
	}

	return c.Writer.Write(b)
}

func (c *ConnAES) Close() error {
	c.SetDeadline(time.Now().Add(DEFAULT_TIMEOUT))
	return c.Conn.Close()
}

func (c *ConnAES) Upgrade(ts int64, iv []byte, compress bool) {
	c.Reader = newAESReader(c.Conn, ts, iv)
	c.Writer = newAESWriter(c.Conn, ts, iv)

	if compress {
		log.Println("conn compression")
		c.snappyReader = snappy.NewReader(c.Reader)
		c.snappyWriter = snappy.NewBufferedWriter(c.Writer)
	}
}

func NewAESConn(conn net.Conn) *ConnAES {
	return &ConnAES{
		Conn:   conn,
		Reader: conn,
		Writer: conn,
	}
}
