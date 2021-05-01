package proxy

import (
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/lijianying10/simpleTcpReverseProxy/pkg/config"
)

// Runtime proxy runtime
type Runtime struct {
	cfg      *config.Config
	cursor   int
	listener net.Listener
}

func NewRuntime(cfg *config.Config) *Runtime {
	return &Runtime{
		cfg:    cfg,
		cursor: 0,
	}
}

// getNextRemoteAddr round robin LB
func (rt *Runtime) getNextRemoteAddr() string {
	defer func() {
		rt.cursor++
	}()
	return rt.cfg.IPList[rt.cursor%len(rt.cfg.IPList)] + ":" + strconv.Itoa(rt.cfg.TargetPort)
}

// Run start a reverse proxy
func (rt *Runtime) Run() {
	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(rt.cfg.Port))
	if err != nil {
		panic(err)
	}
	rt.listener = listener
	for {
		conn, err := listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			}
			log.Println("[ERROR] accepting connection", err)
			continue
		}
		go func() {
			var err error
			var conn2 net.Conn
			// TODO: Does we need retry?
			conn2, err = net.DialTimeout("tcp", rt.getNextRemoteAddr(), 3*time.Second)
			if err != nil {
				log.Println("[ERROR] dialing remote addr", err)
				conn.Close()
				return
			}
			go io.Copy(conn2, conn)
			io.Copy(conn, conn2)
			conn2.Close()
			conn.Close()
		}()
	}
}

// Stop stop a reverse proxy
func (rt *Runtime) Stop() {
	rt.listener.Close()
}
