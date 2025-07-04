package ws

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
	"titan-tunnel/server/model"
	"titan-tunnel/server/socks5"
	"titan-tunnel/server/ws/pb"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type TunOptions struct {
	Id    string
	OS    string
	VMAPI string
	IP    string
	// seconds
	UDPTimeout int
	TCPTimeout int
	// Driver string
}

// Tunnel Tunnel
type Tunnel struct {
	// id        string
	conn      *websocket.Conn
	writeLock sync.Mutex
	waitping  int

	proxys sync.Map

	opts *TunOptions
	// commandList sync.Map
	waitList sync.Map
	tunMgr   *TunnelManager
}

func newTunnel(conn *websocket.Conn, tunMgr *TunnelManager, opts *TunOptions) *Tunnel {

	t := &Tunnel{
		conn:   conn,
		opts:   opts,
		tunMgr: tunMgr,
	}

	conn.SetPingHandler(func(data string) error {
		t.writePong([]byte(data))
		return nil
	})

	conn.SetPongHandler(func(data string) error {
		t.onPong()
		return nil
	})

	return t
}

func (t *Tunnel) writePong(msg []byte) error {
	t.writeLock.Lock()
	defer t.writeLock.Unlock()
	return t.conn.WriteMessage(websocket.PongMessage, msg)
}

func (t *Tunnel) onPong() {
	t.waitping = 0
}

func (t *Tunnel) serve() {
	defer t.close()

	c := t.conn
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			logx.Error("Tunnel read failed:", err)
			return
		}

		err = t.onMessage(message)
		if err != nil {
			logx.Error("Tunnel.serve onMessage failed:", err)
		}
	}
}

func (t *Tunnel) onMessage(data []byte) error {
	logx.Debugf("Tunnel.onMessage")

	msg := &pb.Message{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return err
	}

	switch msg.Type {
	case pb.MessageType_COMMAND:
		// return t.onControlMessage(msg.GetSessionId(), msg.Payload)
	case pb.MessageType_PROXY_SESSION_CREATE:
		return t.onProxySessionCreateReply(msg.GetSessionId(), msg.Payload)
	case pb.MessageType_PROXY_SESSION_DATA:
		return t.onProxySessionDataFromTunnel(msg.GetSessionId(), msg.Payload)
	case pb.MessageType_PROXY_SESSION_CLOSE:
		return t.onProxySessionClose(msg.GetSessionId())
	case pb.MessageType_PROXY_UDP_DATA:
		return t.onProxyUDPDataFromTunnel(msg.GetSessionId(), msg.Payload)
	default:
		logx.Errorf("Tunnel.onMessage, unsupport message type:%d", msg.Type)

	}
	return nil
}

func (t *Tunnel) onProxySessionCreateReply(sessionID string, payload []byte) error {
	logx.Debugf("Tunnel.onProxySessionCreateComplete")
	v, ok := t.waitList.Load(sessionID)
	if !ok || v == nil {
		return fmt.Errorf("Tunnel.onProxySessionCreateComplete not found session:%s", sessionID)
	}

	ch := v.(chan []byte)

	select {
	case ch <- payload:
	default:
		logx.Errorf("Tunnel.onProxySessionCreateComplete: channel full or no listener for session %s", sessionID)
	}
	return nil
}

func (t *Tunnel) onProxySessionClose(sessionID string) error {
	logx.Debugf("Tunnel.onProxySessionClose session id: %s", sessionID)
	v, ok := t.proxys.Load(sessionID)
	if !ok {
		return fmt.Errorf("Tunnel.onProxySessionClose, can not found session %s", sessionID)
	}

	session := v.(*TCPProxy)
	session.closeByCleint()

	return nil
}

func (t *Tunnel) onProxySessionDataFromTunnel(sessionID string, data []byte) error {
	logx.Debugf("Tunnel.onProxySessionDataFromTunnel session id: %s", sessionID)
	v, ok := t.proxys.Load(sessionID)
	if !ok {
		t.onProxyTCPConnClose(sessionID)
		return fmt.Errorf("Tunnel.onProxySessionDataFromTunnel, can not found session %s", sessionID)
	}

	session := v.(*TCPProxy)
	return session.write(data)
}

func (t *Tunnel) onProxyTCPConnClose(sessionID string) {
	logx.Debugf("Tunnel.onProxyConnClose, session id: %s", sessionID)
	msg := &pb.Message{}
	msg.Type = pb.MessageType_PROXY_SESSION_CLOSE
	msg.SessionId = sessionID
	msg.Payload = nil

	buf, err := proto.Marshal(msg)
	if err != nil {
		logx.Errorf("Tunnel.onProxyConnClose, EncodeMessage failed:%s", err.Error())
		return
	}

	if err = t.write(buf); err != nil {
		logx.Errorf("Tunnel.onProxyConnClose, write message to tunnel failed:%s", err.Error())
	}
}

func (t *Tunnel) onProxyDataFromProxy(sessionID string, data []byte) {
	logx.Debugf("Tunnel.onProxyDataFromProxy, data len: %d", len(data))

	msg := &pb.Message{}
	msg.Type = pb.MessageType_PROXY_SESSION_DATA
	msg.SessionId = sessionID
	msg.Payload = data

	buf, err := proto.Marshal(msg)
	if err != nil {
		logx.Errorf("Tunnel.onProxyDataFromProxy proto message failed:%s", err.Error())
		return
	}

	if err = t.write(buf); err != nil {
		logx.Errorf("Tunnel.onProxyDataFromProxy, write message to tunnel failed:%s", err.Error())
	}

	// logx.Debugf("Tunnel.onProxyDataFromProxy write message to tunnel success")
}

func (t *Tunnel) acceptSocks5TCPConn(conn net.Conn, targetInfo *socks5.SocksTargetInfo) error {
	logx.Debugf("acceptSocks5TCPConn, dest %s:%d", targetInfo.DomainName, targetInfo.Port)

	sessionID := uuid.NewString()

	addr := fmt.Sprintf("%s:%d", targetInfo.DomainName, targetInfo.Port)
	err := t.onClientCreateByDomain(&pb.DestAddr{Addr: addr}, sessionID)
	if err != nil {
		return fmt.Errorf("Tunnel.acceptSocks5TCPConn client create by Domain failed, addr:%s, err:%v", addr, err)
	}

	if len(targetInfo.ExtraBytes) > 0 {
		t.onProxyDataFromProxy(sessionID, targetInfo.ExtraBytes)
	}

	proxyTCP := newTCPProxy(sessionID, conn, t)

	t.proxys.Store(sessionID, proxyTCP)
	defer t.proxys.Delete(sessionID)

	return proxyTCP.proxyConn()
}

func (t *Tunnel) onClientCreateByDomain(dest *pb.DestAddr, sessionID string) error {
	logx.Debugf("Tunnel.onClientCreateByDomain, dest %s", dest.Addr)

	buf, err := proto.Marshal(dest)
	if err != nil {
		return err
	}

	msg := &pb.Message{}
	msg.Type = pb.MessageType_PROXY_SESSION_CREATE
	msg.SessionId = sessionID
	msg.Payload = buf

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t.opts.TCPTimeout)*time.Second)
	defer cancel()

	reply, err := t.requestCreateProxySession(ctx, msg)
	if err != nil {
		return err
	}

	if !reply.Success {
		return fmt.Errorf(reply.ErrMsg)
	}

	return nil

}

func (t *Tunnel) requestCreateProxySession(ctx context.Context, in *pb.Message) (*pb.CreateSessionReply, error) {
	data, err := proto.Marshal(in)
	if err != nil {
		return nil, err
	}

	ch := make(chan []byte)
	t.waitList.Store(in.GetSessionId(), ch)
	defer t.waitList.Delete(in.GetSessionId())

	err = t.write(data)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case data := <-ch:
			out := &pb.CreateSessionReply{}
			err = proto.Unmarshal(data, out)
			if err != nil {
				return nil, fmt.Errorf("can not unmarshal replay:%s", err.Error())
			}
			return out, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

}

func (t *Tunnel) onProxyUDPDataFromTunnel(sessionID string, data []byte) error {
	proxy, ok := t.proxys.Load(sessionID)
	if !ok {
		return fmt.Errorf("Tunnel.onProxyUDPDataFromTunnel session %s not exist", sessionID)

	}
	udp := proxy.(*UDPProxy)

	return udp.writeToSrc(data)
}

func (t *Tunnel) acceptSocks5UDPData(conn socks5.UDPConn, udpInfo *socks5.Socks5UDPInfo, data []byte) error {
	sessionID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(fmt.Sprintf("%s%s%s", udpInfo.UDPServerID, udpInfo.Src, udpInfo.Dest))).String()

	var udp *UDPProxy
	v, ok := t.proxys.Load(sessionID)
	if !ok {
		udp = newProxyUDP(sessionID, conn, udpInfo, t, t.opts.UDPTimeout)
		t.proxys.Store(sessionID, udp)

		go udp.waitTimeout()
	} else {
		udp = v.(*UDPProxy)
	}

	return udp.writeToDest(data)
}

func (t *Tunnel) keepalive() {
	if t.waitping > 3 {
		t.conn.Close()
		return
	}

	t.writeLock.Lock()
	defer t.writeLock.Unlock()

	model.SetNodeOnline(t.tunMgr.redis, t.opts.Id)

	b := make([]byte, 8)

	now := time.Now().Unix()
	binary.LittleEndian.PutUint64(b, uint64(now))

	t.conn.WriteMessage(websocket.PingMessage, b)

	t.waitping++
}

func (t *Tunnel) write(msg []byte) error {
	t.writeLock.Lock()
	defer t.writeLock.Unlock()

	return t.conn.WriteMessage(websocket.BinaryMessage, msg)
}

func (t *Tunnel) close() {
	t.proxys.Range(func(key, value any) bool {
		session, ok := value.(*TCPProxy)
		if ok {
			session.close()
		}
		return true
	})
	t.proxys.Clear()
}
