package server

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"
	"titan-vm/proto"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// CtrlTunnel CtrlTunnel
type CtrlTunnel struct {
	id        string
	conn      *websocket.Conn
	writeLock sync.Mutex
	waitping  int

	proxySessions map[string]*ProxySession
}

func newCtrlTunnel(id string, conn *websocket.Conn) *CtrlTunnel {

	ct := &CtrlTunnel{
		id:            id,
		conn:          conn,
		proxySessions: make(map[string]*ProxySession),
	}

	conn.SetPingHandler(func(data string) error {
		ct.writePong([]byte(data))
		return nil
	})

	conn.SetPongHandler(func(data string) error {
		ct.onPong()
		return nil
	})

	return ct
}

func (ct *CtrlTunnel) writePong(msg []byte) error {
	ct.writeLock.Lock()
	defer ct.writeLock.Unlock()
	return ct.conn.WriteMessage(websocket.PongMessage, msg)
}

func (ct *CtrlTunnel) onPong() {
	ct.waitping = 0
}

func (ct *CtrlTunnel) serve() {
	defer ct.onClose()

	c := ct.conn
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Error("CtrlTunnel read failed:", err)
			return
		}

		err = ct.onMessage(message)
		if err != nil {
			log.Error("CtrlTunnel onMessage failed:", err)
		}
	}
}

func (ct *CtrlTunnel) onMessage(data []byte) error {
	log.Debugf("CtrlTunnel.onMessage")

	msg := &proto.Message{}
	err := msg.DecodeMessage(data)
	if err != nil {
		return err
	}

	switch msg.Type {
	case proto.MsgTypeControl:
		return ct.onControlMessage()
	case proto.MsgTypeProxySessionData:
		return ct.onProxySessionData(msg.SessionID, msg.Payload)
	case proto.MsgTypeProxySessionClose:
		return ct.onProxySessionClose(msg.SessionID)
	default:
		log.Printf("onMessage, unsupport message type:%d", msg.Type)

	}
	return nil
}

func (ct *CtrlTunnel) onControlMessage() error {
	log.Debugf("CtrlTunnel.onControlMessage")
	return nil
}

func (ct *CtrlTunnel) onProxySessionClose(sessionID string) error {
	log.Debugf("CtrlTunnel.onProxySessionClose session id: %s", sessionID)
	session := ct.proxySessions[sessionID]
	if session == nil {
		return fmt.Errorf("onProxySessionClose, can not found session %s", sessionID)
	}

	session.close()
	return nil
}

func (ct *CtrlTunnel) onProxySessionData(sessionID string, data []byte) error {
	log.Debugf("CtrlTunnel.onProxySessionData session id: %s", sessionID)
	session := ct.proxySessions[sessionID]
	if session == nil {
		return fmt.Errorf("onProxySessionData, can not found session %s", sessionID)
	}

	return session.write(data)
}

func (ct *CtrlTunnel) onLibvirtClientAcceptRequest(conn *websocket.Conn, dest *proto.DestAddr) error {
	log.Debugf("onLibvirtClientAcceptRequest, dest %s:%s", dest.Network, dest.Addr)

	buf, err := dest.EncodeMessage()
	if err != nil {
		return err
	}

	msg := &proto.Message{}
	msg.Type = proto.MsgTypeProxySessionCreate
	msg.SessionID = uuid.NewString()
	msg.Payload = buf

	buf, err = msg.EncodeMessage()
	if err != nil {
		return err
	}

	if err := ct.write(buf); err != nil {
		return err
	}

	proxySession := &ProxySession{id: msg.SessionID, conn: conn, sessionType: SessionTypeTcp}
	ct.proxySessions[msg.SessionID] = proxySession

	defer delete(ct.proxySessions, msg.SessionID)

	return proxySession.proxyConn(ct)
}

func (ct *CtrlTunnel) onWebVncAcceptRequest(conn *websocket.Conn, dest *proto.DestAddr) error {
	log.Debugf("onWebVncAcceptRequest, dest %s:%s", dest.Network, dest.Addr)

	buf, err := dest.EncodeMessage()
	if err != nil {
		return err
	}

	msg := &proto.Message{}
	msg.Type = proto.MsgTypeProxySessionCreate
	msg.SessionID = uuid.NewString()
	msg.Payload = buf

	buf, err = msg.EncodeMessage()
	if err != nil {
		return err
	}

	ct.write(buf)

	proxySession := &ProxySession{id: msg.SessionID, conn: conn, sessionType: SessionTypeWebsocket}
	ct.proxySessions[msg.SessionID] = proxySession

	defer delete(ct.proxySessions, msg.SessionID)

	return proxySession.proxyConn(ct)
}

func (ct *CtrlTunnel) onProxyConnClose(sessionID string) {
	log.Debugf("onProxyConnClose, session id: %s", sessionID)
	msg := &proto.Message{}
	msg.Type = proto.MsgTypeProxySessionClose
	msg.SessionID = sessionID
	msg.Payload = nil

	buf, err := msg.EncodeMessage()
	if err != nil {
		log.Errorf("CtrlTunnel.onProxyConnClose, EncodeMessage failed:%s", err.Error())
		return
	}

	if err = ct.write(buf); err != nil {
		log.Errorf("CtrlTunnel.onProxyConnClose, write message to tunnel failed:%s", err.Error())
	}
}

func (ct *CtrlTunnel) onProxyData(sessionID string, data []byte) {
	log.Debugf("CtrlTunnel.onProxyData")

	msg := proto.Message{}
	msg.Type = proto.MsgTypeProxySessionData
	msg.SessionID = sessionID
	msg.Payload = data

	buf, err := msg.EncodeMessage()
	if err != nil {
		log.Errorf("onProxyData proto message failed:%s", err.Error())
		return
	}

	if err = ct.write(buf); err != nil {
		log.Errorf("CtrlTunnel.onProxyData, write message to tunnel failed:%s", err.Error())
	}

	log.Debugf("CtrlTunnel.onProxyData write message to tunnel success")
}

func (ct *CtrlTunnel) keepalive() {
	if ct.waitping > 3 {
		ct.conn.Close()
		return
	}

	ct.writeLock.Lock()
	defer ct.writeLock.Unlock()

	b := make([]byte, 8)

	now := time.Now().Unix()
	binary.LittleEndian.PutUint64(b, uint64(now))

	ct.conn.WriteMessage(websocket.PingMessage, b)

	ct.waitping++
}

func (ct *CtrlTunnel) write(msg []byte) error {
	ct.writeLock.Lock()
	defer ct.writeLock.Unlock()

	return ct.conn.WriteMessage(websocket.BinaryMessage, msg)
}

func (ct *CtrlTunnel) onClose() {
	for _, session := range ct.proxySessions {
		session.close()
	}
}
