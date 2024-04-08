package gotun

import (
	"context"
	"io"
	"log"
	"net"
	"runtime/debug"
	"sync"

	"github.com/bytedance/gopkg/lang/mcache"
	"github.com/bytedance/gopkg/util/gopool"
)

type TCPInboundOptions struct {
	Addr       string             `yaml:"addr"`
	TargetAddr string             `yaml:"target_addr"`
	SSHOptions SSHOutboundOptions `yaml:"ssh"`
}

type TCPInbound struct {
	Listener net.Listener
	dialer   *SSH
	Options  TCPInboundOptions
}

func (s *TCPInbound) Run() (err error) {
	if err = s.ListenTCP(); err != nil {
		log.Println("ListenTCP:", err)
		return
	}

	return
}

func (s *TCPInbound) Shutdown() {
	s.Listener.Close()
	s.dialer.Close()
}

func (s *TCPInbound) ListenTCP() (err error) {
	s.dialer, err = NewSSH(context.Background(), s.Options.SSHOptions)
	if err != nil {
		return
	}

	s.Listener, err = net.Listen("tcp", s.Options.Addr)
	if err != nil {
		return
	}

	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("ListenTCP crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
			}
		}()

		for {
			var conn net.Conn
			conn, err = s.Listener.Accept()
			if err != nil {
				log.Printf("accept error , ERR:%s", err)
				break
			}

			gopool.Go(func() {
				defer func() {
					if e := recover(); e != nil {
						log.Printf("connection handler crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
					}
				}()

				s.executeConn(conn)
			})
		}
	}()

	log.Printf("tcp tunnel on %s", s.Options.Addr)

	return
}

func (s *TCPInbound) executeConn(inConn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("tcp conn handler crashed with err : %s \nstack: %s", err, string(debug.Stack()))
		}
	}()

	err := s.OutToTCP(s.Options.TargetAddr, &inConn)

	if err != nil {
		log.Printf("connect to %s fail, ERR:%s", s.Options.TargetAddr, err)
		inConn.Close()
	}
}

func (s *TCPInbound) OutToTCP(address string, inConn *net.Conn) (err error) {
	inAddr := (*inConn).RemoteAddr().String()
	inLocalAddr := (*inConn).LocalAddr().String()

	var outConn net.Conn
	outConn, err = s.dialer.Dial(context.Background(), "tcp", address)

	if err != nil {
		log.Printf("connect to %s, err:%s", address, err)
		(*inConn).Close()
		return
	}

	s.IoBind((*inConn), outConn, func(err error) {
		log.Printf("conn %s - %s released [%s]", inAddr, inLocalAddr, address)

		(*inConn).Close()
		outConn.Close()
	})

	log.Printf("conn %s - %s connected [%s]", inAddr, inLocalAddr, address)
	return
}

func (s *TCPInbound) IoBind(src, dst net.Conn, fnClose func(err error)) {
	var one = &sync.Once{}

	gopool.Go(func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("IoBind crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
			}
		}()

		buf := mcache.Malloc(32 * 1024)
		defer mcache.Free(buf)

		if _, err := io.CopyBuffer(dst, src, buf); err != nil {
			one.Do(func() {
				fnClose(err)
			})
		}
	})

	gopool.Go(func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("IoBind crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
			}
		}()

		buf := mcache.Malloc(32 * 1024)
		defer mcache.Free(buf)

		if _, err := io.CopyBuffer(src, dst, buf); err != nil {
			one.Do(func() {
				fnClose(err)
			})
		}
	})
}

func NewTCPInbound(opts TCPInboundOptions) *TCPInbound {
	inbound := new(TCPInbound)
	inbound.Options = opts
	return inbound
}
