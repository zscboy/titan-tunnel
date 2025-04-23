package client

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
	"titan-vm/proto"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Tunnel struct {
	uuid      string
	conn      *websocket.Conn
	writeLock sync.Mutex

	url      string
	waitping int

	proxySessions sync.Map
}

func NewTunnel(serverUrl, uuid string) (*Tunnel, error) {
	tun := &Tunnel{
		uuid:      uuid,
		writeLock: sync.Mutex{},
		url:       serverUrl,
	}

	return tun, nil
}

func (t *Tunnel) Connect() error {
	url := fmt.Sprintf("%s?uuid=%s", t.url, t.uuid)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("dial %s failed %s", url, err.Error())
	}

	conn.SetPingHandler(func(data string) error {
		t.writePong([]byte(data))
		return nil
	})

	conn.SetPongHandler(func(data string) error {
		t.onPong([]byte(data))
		return nil
	})

	t.conn = conn

	// log.Infof("response header: %#v", resp.Header)
	log.Infof("new tun %s", url)
	return nil
}

func (t *Tunnel) Destroy() error {
	if t.conn != nil {
		return t.conn.Close()
	}

	return nil
}

func (t *Tunnel) Serve() error {
	conn := t.conn
	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Error("Error reading message:", err)
			break
		}

		if messageType != websocket.BinaryMessage {
			log.Errorf("unsupport message type %d", messageType)
			continue
		}

		if err = t.onTunnelMsg(p); err != nil {
			log.Errorf("onTunnelMsg: %s", err.Error())
		}
	}

	log.Debugf("tunnel %s close", t.uuid)

	return nil
}

func (t *Tunnel) onTunnelMsg(message []byte) error {
	msg := &proto.Message{}
	err := msg.DecodeMessage(message)
	if err != nil {
		return fmt.Errorf("DecodeMessage failed:%s", err.Error())
	}

	log.Debugf("Tunnel.onTunnelMsg, msgType:%s, session id:%s", msg.Type.String(), msg.SessionID)

	switch msg.Type {
	case proto.MsgTypeControl:
		return t.onControlMessage(msg)
	case proto.MsgTypeProxySessionCreate:
		return t.onProxySessionCreate(msg)
	case proto.MsgTypeProxySessionData:
		return t.onProxySessionData(msg)
	case proto.MsgTypeProxySessionClose:
		return t.onProxySessionClose(msg)
	default:
		log.Errorf("onTunnelMsg unsupoort message type %d", msg.Type)
	}

	return nil
}

func (t *Tunnel) onControlMessage(_ *proto.Message) error {
	log.Debugf("Tunnel.onControlMessage")
	return nil
}

func (t *Tunnel) onProxySessionCreate(msg *proto.Message) error {
	session, ok := t.proxySessions.Load(msg.SessionID)
	if ok {
		ps := session.(ProxySession)
		ps.close()
		// TODO: wait session delete
	}

	destAddr := &proto.DestAddr{}
	err := destAddr.DecodeMessage(msg.Payload)
	if err != nil {
		return fmt.Errorf("decode message failed:%s", err.Error())
	}

	if destAddr.Network != "unix" && destAddr.Network != "tcp" {
		return fmt.Errorf("dest addr unsupport netowrk type %s", destAddr.Network)
	}

	conn, err := net.DialTimeout(destAddr.Network, destAddr.Addr, 3*time.Second)
	if err != nil {
		return fmt.Errorf("dial %s, network: %s, failed %s", destAddr.Addr, destAddr.Network, err.Error())
	}

	proxySession := ProxySession{id: msg.SessionID, conn: conn}
	t.proxySessions.Store(msg.SessionID, proxySession)

	go proxySession.proxyConn(t)

	return nil
}

func (t *Tunnel) onProxySessionData(msg *proto.Message) error {
	session, ok := t.proxySessions.Load(msg.SessionID)
	if !ok {
		return fmt.Errorf("onProxySessionData session %s not found", msg.SessionID)
	}

	ps := session.(ProxySession)
	return ps.write(msg.Payload)
}

func (t *Tunnel) onProxySessionClose(msg *proto.Message) error {
	session, ok := t.proxySessions.Load(msg.SessionID)
	if !ok {
		return fmt.Errorf("onProxySessionData session %s not found", msg.SessionID)
	}

	ps := session.(ProxySession)
	ps.close()
	return nil
}

func (t *Tunnel) onProxyConnClose(sessionID string) {
	log.Debugf("Tunnel.onProxyConnClose session id:%s", sessionID)
	msg := &proto.Message{}
	msg.Type = proto.MsgTypeProxySessionClose
	msg.SessionID = sessionID
	msg.Payload = nil

	buf, err := msg.EncodeMessage()
	if err != nil {
		log.Errorf("onProxyData encode message failed:%s", err.Error())
		return
	}

	if err = t.write(buf); err != nil {
		log.Errorf("write message to tunnel failed:%s", err.Error())
	}
}

func (t *Tunnel) onProxyData(sessionID string, data []byte) {
	log.Debugf("Tunnel.onProxyData session id:%s", sessionID)
	msg := &proto.Message{}
	msg.Type = proto.MsgTypeProxySessionData
	msg.SessionID = sessionID
	msg.Payload = data

	buf, err := msg.EncodeMessage()
	if err != nil {
		log.Errorf("onProxyData encode message failed:%s", err.Error())
		return
	}

	if err = t.write(buf); err != nil {
		log.Errorf("write message to tunnel failed:%s", err.Error())
	}
}

func (t *Tunnel) writePong(msg []byte) error {
	t.writeLock.Lock()
	defer t.writeLock.Unlock()
	return t.conn.WriteMessage(websocket.PongMessage, msg)
}

func (t *Tunnel) onPong(_ []byte) {
	t.waitping = 0
}

func (t *Tunnel) write(msg []byte) error {
	t.writeLock.Lock()
	defer t.writeLock.Unlock()
	return t.conn.WriteMessage(websocket.BinaryMessage, msg)
}
