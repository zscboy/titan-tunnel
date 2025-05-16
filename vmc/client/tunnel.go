package client

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"sync"
	"time"
	"titan-vm/pb"
	"titan-vm/vmc/downloader"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type Tunnel struct {
	uuid      string
	conn      *websocket.Conn
	writeLock sync.Mutex

	url      string
	waitping int

	proxySessions sync.Map
	isDestroy     bool
	command       *Command
	vmapi         string
}

func NewTunnel(serverUrl, uuid, vmapi string) (*Tunnel, error) {
	tun := &Tunnel{
		uuid:      uuid,
		writeLock: sync.Mutex{},
		url:       serverUrl,
		isDestroy: false,
		vmapi:     vmapi,
	}

	tun.command = &Command{tunnel: tun, downloadManager: downloader.NewManager()}

	return tun, nil
}

func (t *Tunnel) Connect() error {
	url := fmt.Sprintf("%s?uuid=%s&os=%s&vmapi=%s", t.url, t.uuid, runtime.GOOS, t.vmapi)

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
	msg := &pb.Message{}
	err := proto.Unmarshal(message, msg)
	if err != nil {
		return fmt.Errorf("DecodeMessage failed:%s", err.Error())
	}

	log.Debugf("Tunnel.onTunnelMsg, msgType:%s, session id:%s", msg.Type.String(), msg.GetSessionId())

	switch msg.Type {
	case pb.MessageType_CONTROL:
		return t.onControlMessage(msg)
	case pb.MessageType_PROXY_SESSION_CREATE:
		return t.onProxySessionCreate(msg)
	case pb.MessageType_PROXY_SESSION_DATA:
		return t.onProxySessionData(msg)
	case pb.MessageType_PROXY_SESSION_CLOSE:
		return t.onProxySessionClose(msg)
	default:
		log.Errorf("onTunnelMsg unsupoort message type %d", msg.Type)
	}

	return nil
}

func (t *Tunnel) onControlMessage(msg *pb.Message) error {
	cmd := &pb.Command{}
	err := proto.Unmarshal(msg.GetPayload(), cmd)
	if err != nil {
		return err
	}

	log.Debugf("Tunnel.onControlMessage, cmd type:%s", cmd.GetType().String())

	switch cmd.GetType() {
	case pb.CommandType_DownloadImage:
		return t.command.downloadImage(msg.GetSessionId(), cmd.GetData())
	}
	return nil
}

func (t *Tunnel) onProxySessionCreate(msg *pb.Message) error {
	session, ok := t.proxySessions.Load(msg.GetSessionId())
	if ok {
		ps := session.(ProxySession)
		ps.close()
		// TODO: wait session delete
	}

	destAddr := &pb.DestAddr{}
	err := proto.Unmarshal(msg.GetPayload(), destAddr)
	if err != nil {
		return fmt.Errorf("decode message failed:%s", err.Error())
	}

	if destAddr.Network != "unix" && destAddr.Network != "tcp" {
		return fmt.Errorf("dest addr unsupport netowrk type %s", destAddr.Network)
	}

	conn, err := net.DialTimeout(destAddr.GetNetwork(), destAddr.GetAddr(), 3*time.Second)
	if err != nil {
		t.onProxyConnClose(msg.GetSessionId())
		return fmt.Errorf("dial %s, network: %s, failed %s", destAddr.Addr, destAddr.Network, err.Error())
	}

	proxySession := ProxySession{id: msg.GetSessionId(), conn: conn}
	t.proxySessions.Store(msg.GetSessionId(), proxySession)

	go proxySession.proxyConn(t)

	return nil
}

func (t *Tunnel) onProxySessionData(msg *pb.Message) error {
	session, ok := t.proxySessions.Load(msg.GetSessionId())
	if !ok {
		return fmt.Errorf("onProxySessionData session %s not found", msg.GetSessionId())
	}

	ps := session.(ProxySession)
	return ps.write(msg.Payload)
}

func (t *Tunnel) onProxySessionClose(msg *pb.Message) error {
	session, ok := t.proxySessions.Load(msg.GetSessionId())
	if !ok {
		return fmt.Errorf("onProxySessionData session %s not found", msg.GetSessionId())
	}

	ps := session.(ProxySession)
	ps.close()
	return nil
}

func (t *Tunnel) onProxyConnClose(sessionID string) {
	log.Debugf("Tunnel.onProxyConnClose session id:%s", sessionID)
	msg := &pb.Message{}
	msg.Type = pb.MessageType_PROXY_SESSION_CLOSE
	msg.SessionId = sessionID
	msg.Payload = nil

	buf, err := proto.Marshal(msg)
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
	msg := &pb.Message{}
	msg.Type = pb.MessageType_PROXY_SESSION_DATA
	msg.SessionId = sessionID
	msg.Payload = data

	buf, err := proto.Marshal(msg)
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
