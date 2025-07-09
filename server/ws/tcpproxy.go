package ws

import (
	"fmt"
	"net"

	"github.com/zeromicro/go-zero/core/logx"
)

type TCPProxy struct {
	id              string
	conn            net.Conn
	tunnel          *Tunnel
	userName        string
	isCloseByClient bool
}

func newTCPProxy(id string, conn net.Conn, t *Tunnel, userName string) *TCPProxy {
	return &TCPProxy{id: id, conn: conn, tunnel: t, userName: userName}
}

func (proxy *TCPProxy) close() {
	if proxy.conn == nil {
		logx.Errorf("session %s conn == nil", proxy.id)
		return
	}

	proxy.conn.Close()
}

func (proxy *TCPProxy) closeByCleint() {
	proxy.isCloseByClient = true
	proxy.close()
}

func (proxy *TCPProxy) write(data []byte) error {
	if proxy.conn == nil {
		return fmt.Errorf("session %s conn == nil", proxy.id)
	}

	proxy.tunnel.tunMgr.traffic(proxy.userName, int64(len(data)))
	_, err := proxy.conn.Write(data)
	return err
}

func (proxy *TCPProxy) proxyConn() error {
	conn := proxy.conn
	defer conn.Close()

	// netConn := conn.NetConn()
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			logx.Infof("proxy.proxyConn: %v", err)
			if !proxy.isCloseByClient {
				proxy.tunnel.onProxySessionClose(proxy.id)
			}
			return nil
		}

		proxy.tunnel.tunMgr.traffic(proxy.userName, int64(n))
		proxy.tunnel.onProxyDataFromProxy(proxy.id, buf[:n])
	}
}
