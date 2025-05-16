package client

import (
	"fmt"
	"io"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

const networkErrorCloseByRemoteHost = "An existing connection was forcibly closed by the remote host"

type ProxySession struct {
	id   string
	conn net.Conn
}

func (ps *ProxySession) close() {
	if ps.conn == nil {
		log.Errorf("session %s conn == nil", ps.id)
		return
	}

	ps.conn.Close()
}

func (ps *ProxySession) write(data []byte) error {
	if ps.conn == nil {
		return fmt.Errorf("session %s conn == nil", ps.id)
	}

	_, err := ps.conn.Write(data)
	return err
}

func (ps *ProxySession) proxyConn(t *Tunnel) {
	defer t.proxySessions.Delete(ps.id)

	conn := ps.conn
	defer conn.Close()

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Debugf("serveProxyConn: %s", err.Error())
			if err == io.EOF || strings.Contains(err.Error(), networkErrorCloseByRemoteHost) {
				t.onProxyConnClose(ps.id)
			}
			return
		}
		t.onProxyData(ps.id, buf[:n])
	}
}
