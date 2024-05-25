package tunnel

import (
	"io"
	"log"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/gopkg/lang/mcache"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/taodev/gotun/option"
)

type TunnelTCP struct {
	protocol       string
	tag            string
	addr           string
	password       string
	inboundAuth    bool
	targetAddr     string
	targetPassword string
	outboundAuth   bool

	compression bool

	tcpListener net.Listener
	connCount   int32
}

func (t *TunnelTCP) Type() string {
	return t.protocol
}

func (t *TunnelTCP) Tag() string {
	return t.tag
}

func (t TunnelTCP) Start() (err error) {
	if err = t.ListenTCP(); err != nil {
		return
	}

	return
}

func (t *TunnelTCP) ListenTCP() (err error) {
	t.tcpListener, err = net.Listen("tcp", t.addr)
	if err != nil {
		return
	}

	go t.loopTCP()

	log.Printf("tunnel %s listening on %s", t.protocol, t.addr)

	return
}

func (t *TunnelTCP) loopTCP() {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("loopTCP crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
		}
	}()

	gopool.Go(func() {
		tk := time.NewTicker(10 * time.Second)

		for {
			<-tk.C

			log.Printf("tunnel %s conn count : %d", t.protocol, atomic.LoadInt32(&t.connCount))
		}
	})

	var err error
	for {
		var conn net.Conn
		conn, err = t.tcpListener.Accept()
		if err != nil {
			log.Printf("accept error , ERR:%s", err)
			return
		}

		gopool.Go(func() {
			defer func() {
				if e := recover(); e != nil {
					log.Printf("connection handler crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
				}
			}()

			gopool.Go(func() {
				t.executeConn(conn)
			})
		})
	}
}

func (t *TunnelTCP) executeConn(inConn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("tcp conn handler crashed with err : %s \nstack: %s", err, string(debug.Stack()))
		}
	}()

	in := NewAESConn(inConn)

	// 需要认证
	if t.inboundAuth {
		ts, iv, err := authVerify(inConn, t.password)
		if err != nil {
			log.Println("conn accept auth failed", err)
			inConn.Close()
			return
		}

		in.Upgrade(ts, iv, t.compression)
	}

	// 连接目标
	var outConn net.Conn
	outConn, err := net.Dial("tcp", t.targetAddr)
	if err != nil {
		log.Printf("connect to %s fail, ERR:%s", t.targetAddr, err)
		in.Close()
		return
	}

	log.Printf("dial %s success", t.targetAddr)

	out := NewAESConn(outConn)

	if t.outboundAuth {
		ts, iv, err := auth(outConn, t.targetPassword)
		if err != nil {
			outConn.Close()
			in.Close()
			return
		}

		out.Upgrade(ts, iv, t.compression)
	}

	atomic.AddInt32(&t.connCount, 1)

	t.IoBind(in, out, func(err error) {
		log.Printf("conn %s - %s released [%s]", inConn.RemoteAddr().String(), inConn.LocalAddr().String(), outConn.RemoteAddr().String())
		in.Close()
		out.Close()
		atomic.AddInt32(&t.connCount, -1)
	})

	log.Printf("conn %s - %s connected [%s]", inConn.RemoteAddr().String(), inConn.LocalAddr().String(), outConn.RemoteAddr().String())
}

func (t *TunnelTCP) IoBind(in, out net.Conn, fnClose func(err error)) {
	var once = &sync.Once{}

	gopool.Go(func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("IoBind crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
			}
		}()

		buf := mcache.Malloc(32 * 1024)
		defer mcache.Free(buf)

		_, err := io.CopyBuffer(in, out, buf)
		once.Do(func() {
			fnClose(err)
		})
	})

	gopool.Go(func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("IoBind crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
			}
		}()

		buf := mcache.Malloc(32 * 1024)
		defer mcache.Free(buf)

		_, err := io.CopyBuffer(out, in, buf)
		once.Do(func() {
			fnClose(err)
		})
	})
}

func NewTunnelTCP(options option.Tunnel) *TunnelTCP {
	log.Println(options)
	return &TunnelTCP{
		protocol:       options.Type,
		tag:            options.Tag,
		addr:           options.Addr,
		password:       options.Password,
		inboundAuth:    len(options.Password) > 0,
		targetAddr:     options.TargetAddr,
		targetPassword: options.TargetPassword,
		outboundAuth:   len(options.TargetPassword) > 0,
		compression:    options.Compression,
	}
}
