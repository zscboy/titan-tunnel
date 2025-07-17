package tunnel

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"
	"titan-tunnel/server/api/ws/pb"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

const (
	keepaliveInterval = 10
	waitPoneTimeout   = 3
)

type Pop struct {
	URL   string `json:"server_url"`
	Token string `json:"access_token"`
}

type Tunnel struct {
	uuid      string
	conn      *websocket.Conn
	writeLock sync.Mutex

	url      string
	waitpone int

	proxySessions sync.Map
	proxyUDPs     sync.Map
	isDestroy     bool
	// secondes
	udpTimeout int
	tcpTimeout int
	ctx        context.Context
	ctxCancel  context.CancelFunc
}

func NewTunnel(serverUrl, uuid string, udpTimeout, tcpTimeout int) (*Tunnel, error) {
	tun := &Tunnel{
		uuid:       uuid,
		writeLock:  sync.Mutex{},
		url:        serverUrl,
		isDestroy:  false,
		udpTimeout: udpTimeout,
		tcpTimeout: tcpTimeout,
	}

	return tun, nil
}

func (t *Tunnel) Connect() error {
	serverURL := fmt.Sprintf("%s?nodeid=%s", t.url, t.uuid)
	pop, err := t.getPop(serverURL)
	if err != nil {
		return fmt.Errorf("get pop %s, failed:%v", serverURL, err)
	}

	header := http.Header{}
	header.Add("Authorization", "Bearer "+pop.Token)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t.tcpTimeout)*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s?id=%s&os=%s", pop.URL, t.uuid, runtime.GOOS)
	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, url, header)
	if err != nil {
		var data []byte
		if resp != nil {
			data, _ = io.ReadAll(resp.Body)
		}
		return fmt.Errorf("dial %s failed %s, msg:%s", url, err.Error(), string(data))
	}

	conn.SetPingHandler(func(data string) error {
		t.writePong([]byte(data))
		return nil
	})

	conn.SetPongHandler(func(data string) error {
		t.onPong([]byte(data))
		return nil
	})

	t.ctx, t.ctxCancel = context.WithCancel(context.Background())

	t.waitpone = 0
	t.conn = conn

	go t.keepalive()

	logx.Infof("new tun %s", url)
	return nil
}

func (t *Tunnel) getPop(serverURL string) (*Pop, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	pop := &Pop{}
	err = json.Unmarshal(body, pop)
	if err != nil {
		return nil, err
	}

	return pop, nil
}

func (t *Tunnel) Destroy() error {
	if t.conn != nil {
		t.isDestroy = true
		return t.conn.Close()
	}

	return nil
}

func (t *Tunnel) IsDestroy() bool {
	return t.isDestroy
}

func (t *Tunnel) Serve() error {
	conn := t.conn
	defer t.ctxCancel()
	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			logx.Error("Error reading message:", err)
			break
		}

		if messageType != websocket.BinaryMessage {
			logx.Errorf("unsupport message type %d", messageType)
			continue
		}

		start := time.Now()
		if err = t.onTunnelMsg(p); err != nil {
			logx.Errorf("onTunnelMsg: %s", err.Error())
		}
		logx.Debugf("handle msg cost time: %dms", time.Since(start).Milliseconds())
	}

	logx.Debugf("tunnel %s close", t.uuid)
	t.conn = nil
	return nil
}

func (t *Tunnel) onTunnelMsg(message []byte) error {
	msg := &pb.Message{}
	err := proto.Unmarshal(message, msg)
	if err != nil {
		return fmt.Errorf("DecodeMessage failed:%s", err.Error())
	}

	logx.Debugf("Tunnel.onTunnelMsg, msgType:%s, session id:%s", msg.Type.String(), msg.GetSessionId())

	switch msg.Type {
	case pb.MessageType_PROXY_SESSION_CREATE:
		return t.onProxySessionCreate(msg)
	case pb.MessageType_PROXY_SESSION_DATA:
		return t.onProxySessionDataFromTunnel(msg)
	case pb.MessageType_PROXY_SESSION_CLOSE:
		return t.onProxySessionClose(msg)
	case pb.MessageType_PROXY_UDP_DATA:
		return t.onProxyUDPDataFromTunnel(msg)
	default:
		logx.Errorf("onTunnelMsg unsupoort message type %d", msg.Type)
	}

	return nil
}

func (t *Tunnel) onProxySessionCreate(msg *pb.Message) error {
	go t.createProxySession(msg)

	return nil
}

func (t *Tunnel) createProxySession(msg *pb.Message) error {
	_, ok := t.proxySessions.Load(msg.GetSessionId())
	if ok {
		return t.createProxySessionReply(msg.GetSessionId(), nil)
	}

	destAddr := &pb.DestAddr{}
	err := proto.Unmarshal(msg.GetPayload(), destAddr)
	if err != nil {
		return t.createProxySessionReply(msg.GetSessionId(), err)
	}

	conn, err := net.DialTimeout("tcp", destAddr.GetAddr(), time.Duration(t.tcpTimeout)*time.Second)
	if err != nil {
		logx.Errorf("dial %s, failed %s", destAddr.Addr, err.Error())
		return t.createProxySessionReply(msg.GetSessionId(), err)
	}

	logx.Debugf("Tunnel.onProxySessionCreate new proxy dest %s", destAddr.Addr)

	proxySession := &TCPProxy{id: msg.GetSessionId(), conn: conn}
	t.proxySessions.Store(msg.GetSessionId(), proxySession)

	t.createProxySessionReply(msg.GetSessionId(), nil)

	proxySession.proxyConn(t)

	return nil
}

func (t *Tunnel) createProxySessionReply(sessionID string, err error) error {
	reply := &pb.CreateSessionReply{Success: true}
	if err != nil {
		reply.Success = false
		reply.ErrMsg = err.Error()
	}

	buf, err := proto.Marshal(reply)
	if err != nil {
		return err
	}

	msg := &pb.Message{}
	msg.Type = pb.MessageType_PROXY_SESSION_CREATE
	msg.SessionId = sessionID
	msg.Payload = buf

	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return t.write(data)
}

func (t *Tunnel) onProxySessionDataFromTunnel(msg *pb.Message) error {
	session, ok := t.proxySessions.Load(msg.GetSessionId())
	if !ok {
		return fmt.Errorf("onProxySessionDataFromTunnel session %s not found", msg.GetSessionId())
	}

	proxy := session.(*TCPProxy)
	return proxy.write(msg.Payload)
}

func (t *Tunnel) onProxySessionClose(msg *pb.Message) error {
	session, ok := t.proxySessions.Load(msg.GetSessionId())
	if !ok {
		return fmt.Errorf("onProxySessionClose session %s not found", msg.GetSessionId())
	}

	proxy := session.(*TCPProxy)
	proxy.closeByServer()
	return nil
}

func (t *Tunnel) onProxyConnClose(sessionID string) {
	logx.Debugf("Tunnel.onProxyConnClose session id:%s", sessionID)
	msg := &pb.Message{}
	msg.Type = pb.MessageType_PROXY_SESSION_CLOSE
	msg.SessionId = sessionID
	msg.Payload = nil

	buf, err := proto.Marshal(msg)
	if err != nil {
		logx.Errorf("onProxyData encode message failed:%s", err.Error())
		return
	}

	if err = t.write(buf); err != nil {
		logx.Errorf("write message to tunnel failed:%s", err.Error())
	}
}

func (t *Tunnel) onProxyUDPDataFromTunnel(msg *pb.Message) error {
	udpData := &pb.UDPData{}
	if err := proto.Unmarshal(msg.Payload, udpData); err != nil {
		return err
	}

	id := msg.SessionId
	proxy, ok := t.proxyUDPs.Load(id)
	if ok {
		proxyUDP := proxy.(*UDPProxy)
		return proxyUDP.write(udpData.GetData())
	}

	raddr, err := net.ResolveUDPAddr("udp", udpData.Addr)
	if err != nil {
		logx.Error("tunnel.onProxyUDPDataFromTunnel, ResolveUDPAddr failed:", err)
		return err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		fmt.Println("tunnel.onProxyUDPDataFromTunnel, DialUDP", err)
		return err
	}

	proxyUDP := &UDPProxy{id: id, conn: conn, timeout: t.udpTimeout}
	if err := proxyUDP.write(udpData.GetData()); err != nil {
		return err
	}

	t.proxyUDPs.Store(id, proxyUDP)

	go proxyUDP.serve(t)

	return nil
}

// data from proxy
func (t *Tunnel) onProxySessionDataFromProxy(sessionID string, data []byte) {
	logx.Debugf("Tunnel.onProxySessionDataFromProxy session id:%s", sessionID)
	msg := &pb.Message{}
	msg.Type = pb.MessageType_PROXY_SESSION_DATA
	msg.SessionId = sessionID
	msg.Payload = data

	buf, err := proto.Marshal(msg)
	if err != nil {
		logx.Errorf("onProxyData encode message failed:%s", err.Error())
		return
	}

	if err = t.write(buf); err != nil {
		logx.Errorf("write message to tunnel failed:%s", err.Error())
	}
}

func (t *Tunnel) onProxyUdpDataFromProxy(sessionID string, data []byte) {
	logx.Debugf("Tunnel.onProxySessionDataFromProxy session id:%s", sessionID)
	msg := &pb.Message{}
	msg.Type = pb.MessageType_PROXY_UDP_DATA
	msg.SessionId = sessionID
	msg.Payload = data

	buf, err := proto.Marshal(msg)
	if err != nil {
		logx.Errorf("onProxyData encode message failed:%s", err.Error())
		return
	}

	if err = t.write(buf); err != nil {
		logx.Errorf("write message to tunnel failed:%s", err.Error())
	}
}

func (t *Tunnel) writePong(msg []byte) error {
	t.writeLock.Lock()
	defer t.writeLock.Unlock()
	return t.conn.WriteMessage(websocket.PongMessage, msg)
}

func (t *Tunnel) writePing(msg []byte) error {
	t.writeLock.Lock()
	defer t.writeLock.Unlock()
	return t.conn.WriteMessage(websocket.PingMessage, msg)
}

func (t *Tunnel) onPong(_ []byte) {
	t.waitpone = 0
}

func (t *Tunnel) write(msg []byte) error {
	t.writeLock.Lock()
	defer t.writeLock.Unlock()
	return t.conn.WriteMessage(websocket.BinaryMessage, msg)
}

func (t *Tunnel) keepalive() {
	ticker := time.NewTicker(keepaliveInterval * time.Second)
	defer ticker.Stop()
	defer logx.Debug("keepalive exit")

	for {
		select {
		case <-ticker.C:
			logx.Debug("keepalive tick")
			if t.conn == nil {
				return
			}

			if t.waitpone > waitPoneTimeout {
				t.conn.Close()
			} else {
				t.waitpone++

				b := make([]byte, 8)
				now := time.Now().Unix()
				binary.LittleEndian.PutUint64(b, uint64(now))

				t.writePing(b)
			}
		case <-t.ctx.Done():
			logx.Debug("keepalive stopped")
			return

		}
	}
}
