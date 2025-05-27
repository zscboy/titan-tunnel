package ws

import (
	"fmt"
	"io"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const networkErrorCloseByRemoteHost = "An existing connection was forcibly closed by the remote host"

type TransportType uint16

const (
	TransportTypeUnknown TransportType = iota
	TransportTypeTcp
	TransportTypeWebsocket
)

type ProxySession struct {
	id            string
	conn          *websocket.Conn
	transportType TransportType
}

func (ps *ProxySession) close() {
	if ps.conn == nil {
		logx.Errorf("session %s conn == nil", ps.id)
		return
	}

	ps.conn.Close()
}

func (ps *ProxySession) write(data []byte) error {
	if ps.conn == nil {
		return fmt.Errorf("session %s conn == nil", ps.id)
	}

	switch ps.transportType {
	case TransportTypeWebsocket:
		return ps.conn.WriteMessage(websocket.BinaryMessage, data)
	case TransportTypeTcp:
		conn := ps.conn.NetConn()
		_, err := conn.Write(data)
		return err
	default:
		return fmt.Errorf("unsupport sessionType:%d", ps.transportType)
	}
}

func (ps *ProxySession) proxyConn(ct *CtrlTunnel) error {
	if ps.transportType == TransportTypeWebsocket {
		ps.proxyWebsocketConn(ct)
	} else if ps.transportType == TransportTypeTcp {
		ps.proxyRawConn(ct)
	} else {
		logx.Errorf("proxyConn unsupport sessionType %d", ps.transportType)
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
			logx.Infof("proxyRawConn: %s", err.Error())
			if err == io.EOF || strings.Contains(err.Error(), networkErrorCloseByRemoteHost) {
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
			logx.Debugf("proxyWebsocketConn: %s", err.Error())
			if err == io.EOF || strings.Contains(err.Error(), networkErrorCloseByRemoteHost) {
				ct.onProxyConnClose(ps.id)
			}
			return
		}
		ct.onProxyData(ps.id, p)
	}
}
