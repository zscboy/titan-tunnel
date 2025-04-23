package server

import (
	"fmt"
	"io"
	"strings"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const networkErrorCloseByRemoteHost = "An existing connection was forcibly closed by the remote host"

type SessionType uint16

const (
	SessionTypeUnknown SessionType = iota
	SessionTypeTcp
	SessionTypeWebsocket
)

type ProxySession struct {
	id          string
	conn        *websocket.Conn
	sessionType SessionType
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

	switch ps.sessionType {
	case SessionTypeWebsocket:
		return ps.conn.WriteMessage(websocket.BinaryMessage, data)
	case SessionTypeTcp:
		conn := ps.conn.NetConn()
		_, err := conn.Write(data)
		return err
	default:
		return fmt.Errorf("unsupport sessionType:%d", ps.sessionType)
	}
}

func (ps *ProxySession) proxyConn(ct *CtrlTunnel) error {
	if ps.sessionType == SessionTypeWebsocket {
		ps.proxyWebsocketConn(ct)
	} else if ps.sessionType == SessionTypeTcp {
		ps.proxyRawConn(ct)
	} else {
		log.Errorf("proxyConn unsupport sessionType %d", ps.sessionType)
	}
	return nil
}

func (ps *ProxySession) proxyRawConn(ct *CtrlTunnel) {
	conn := ps.conn
	defer conn.Close()

	netConn := conn.NetConn()
	buf := make([]byte, 4096)
	for {
		n, err := netConn.Read(buf)
		if err != nil {
			log.Debugf("serveProxyConn: %s", err.Error())
			if err == io.EOF {
				ct.onProxyConnClose(ps.id)
			}
			return
		}
		ct.onProxyData(ps.id, buf[:n])
	}
}

func (ps *ProxySession) proxyWebsocketConn(ct *CtrlTunnel) {
	conn := ps.conn
	defer conn.Close()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Debugf("serveProxyConn: %s", err.Error())
			if err == io.EOF || strings.Contains(err.Error(), networkErrorCloseByRemoteHost) {
				ct.onProxyConnClose(ps.id)
			}
			return
		}
		ct.onProxyData(ps.id, p)
	}
}
