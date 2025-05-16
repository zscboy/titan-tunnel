package ws

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
	pb "titan-vm/pb"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type TunOptions struct {
	OS    string
	VMAPI string
	// Driver string
}

// CtrlTunnel CtrlTunnel
type CtrlTunnel struct {
	id        string
	conn      *websocket.Conn
	writeLock sync.Mutex
	waitping  int

	proxySessions sync.Map

	opts        *TunOptions
	commandList sync.Map
}

func newCtrlTunnel(id string, conn *websocket.Conn, opts *TunOptions) *CtrlTunnel {

	ct := &CtrlTunnel{
		id:   id,
		conn: conn,
		opts: opts,
		// proxySessions: make(map[string]*ProxySession),
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
	defer ct.close()

	c := ct.conn
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			logx.Error("CtrlTunnel read failed:", err)
			return
		}

		err = ct.onMessage(message)
		if err != nil {
			logx.Error("CtrlTunnel onMessage failed:", err)
		}
	}
}

func (ct *CtrlTunnel) onMessage(data []byte) error {
	logx.Debugf("CtrlTunnel.onMessage")

	msg := &pb.Message{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return err
	}

	switch msg.Type {
	case pb.MessageType_CONTROL:
		return ct.onControlMessage(msg.GetSessionId(), msg.Payload)
	case pb.MessageType_PROXY_SESSION_DATA:
		return ct.onProxySessionData(msg.GetSessionId(), msg.Payload)
	case pb.MessageType_PROXY_SESSION_CLOSE:
		return ct.onProxySessionClose(msg.GetSessionId())
	default:
		logx.Errorf("onMessage, unsupport message type:%d", msg.Type)

	}
	return nil
}

func (ct *CtrlTunnel) onControlMessage(sessionID string, payload []byte) error {
	logx.Debugf("CtrlTunnel.onControlMessage")
	v, ok := ct.commandList.Load(sessionID)
	if !ok || v == nil {
		return fmt.Errorf("not found command message, id:%s", sessionID)
	}

	ch := v.(chan []byte)

	select {
	case ch <- payload:
	default:
		logx.Errorf("CtrlTunnel.onControlMessage: channel full or no listener for session %s", sessionID)
	}
	return nil
}

func (ct *CtrlTunnel) onProxySessionClose(sessionID string) error {
	logx.Debugf("CtrlTunnel.onProxySessionClose session id: %s", sessionID)
	v, ok := ct.proxySessions.Load(sessionID)
	if !ok {
		return fmt.Errorf("onProxySessionClose, can not found session %s", sessionID)
	}

	session := v.(*ProxySession)
	session.close()

	return nil
}

func (ct *CtrlTunnel) onProxySessionData(sessionID string, data []byte) error {
	logx.Debugf("CtrlTunnel.onProxySessionData session id: %s", sessionID)
	v, ok := ct.proxySessions.Load(sessionID)
	if !ok {
		ct.onProxyConnClose(sessionID)
		return fmt.Errorf("onProxySessionData, can not found session %s", sessionID)
	}

	session := v.(*ProxySession)
	return session.write(data)
}

func (ct *CtrlTunnel) onVmClientAcceptRequest(conn *websocket.Conn, dest *pb.DestAddr, transportType TransportType) error {
	logx.Debugf("onLibvirtClientAcceptRequest, dest %s:%s", dest.Network, dest.Addr)

	buf, err := proto.Marshal(dest)
	if err != nil {
		return err
	}

	sessionID := uuid.NewString()

	msg := &pb.Message{}
	msg.Type = pb.MessageType_PROXY_SESSION_CREATE
	msg.SessionId = sessionID
	msg.Payload = buf

	buf, err = proto.Marshal(msg)
	if err != nil {
		return err
	}

	if err := ct.write(buf); err != nil {
		return err
	}

	proxySession := &ProxySession{id: sessionID, conn: conn, transportType: transportType}
	ct.proxySessions.Store(sessionID, proxySession)

	defer ct.proxySessions.Delete(sessionID)

	return proxySession.proxyConn(ct)
}

func (ct *CtrlTunnel) onProxyConnClose(sessionID string) {
	logx.Debugf("onProxyConnClose, session id: %s", sessionID)
	msg := &pb.Message{}
	msg.Type = pb.MessageType_PROXY_SESSION_CLOSE
	msg.SessionId = sessionID
	msg.Payload = nil

	buf, err := proto.Marshal(msg)
	if err != nil {
		logx.Errorf("CtrlTunnel.onProxyConnClose, EncodeMessage failed:%s", err.Error())
		return
	}

	if err = ct.write(buf); err != nil {
		logx.Errorf("CtrlTunnel.onProxyConnClose, write message to tunnel failed:%s", err.Error())
	}
}

func (ct *CtrlTunnel) onProxyData(sessionID string, data []byte) {
	logx.Debugf("CtrlTunnel.onProxyData, data len: %d", len(data))

	msg := &pb.Message{}
	msg.Type = pb.MessageType_PROXY_SESSION_DATA
	msg.SessionId = sessionID
	msg.Payload = data

	buf, err := proto.Marshal(msg)
	if err != nil {
		logx.Errorf("onProxyData proto message failed:%s", err.Error())
		return
	}

	if err = ct.write(buf); err != nil {
		logx.Errorf("CtrlTunnel.onProxyData, write message to tunnel failed:%s", err.Error())
	}

	logx.Debugf("CtrlTunnel.onProxyData write message to tunnel success")
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

// sendCommand if isWaitReply=true, must set timeout on ctx.
func (ct *CtrlTunnel) sendCommand(ctx context.Context, in *pb.Command, out proto.Message) error {
	bytes, err := proto.Marshal(in)
	if err != nil {
		return err
	}

	msg := &pb.Message{Type: pb.MessageType_CONTROL, SessionId: uuid.NewString(), Payload: bytes}
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	if out == nil {
		return ct.write(data)
	}

	ch := make(chan []byte)
	ct.commandList.Store(msg.GetSessionId(), ch)
	defer ct.commandList.Delete(msg.GetSessionId())

	err = ct.write(data)
	if err != nil {
		return err
	}

	for {
		select {
		case data := <-ch:
			err = proto.Unmarshal(data, out)
			if err != nil {
				return fmt.Errorf("can not unmarshal replay:%s", err.Error())
			}
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

}

func (ct *CtrlTunnel) close() {
	ct.proxySessions.Range(func(key, value any) bool {
		session := value.(*ProxySession)
		session.close()
		return true
	})
	ct.proxySessions.Clear()
}
