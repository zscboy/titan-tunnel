package httpproxy

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"titan-tunnel/server/api/model"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	uuidLength = 16
)

type TunOptions struct {
	Id string
	OS string
	IP string
}

// Tunnel Tunnel
type Tunnel struct {
	conn      *websocket.Conn
	writeLock sync.Mutex
	waitping  int
	opts      *TunOptions
	reqs      sync.Map
	tunMgr    *TunnelManager
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

func (t *Tunnel) writePing(msg []byte) error {
	t.writeLock.Lock()
	defer t.writeLock.Unlock()
	return t.conn.WriteMessage(websocket.PingMessage, msg)
}

func (t *Tunnel) onPong() {
	t.waitping = 0
	model.SetBrowserOnline(t.tunMgr.redis, t.opts.Id)
}

func (t *Tunnel) serve() {
	defer t.close()

	c := t.conn
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			logx.Error("Tunnel read failed:", err)
			return
		}

		if messageType == websocket.TextMessage {
			if err := t.onTextMessage(message); err != nil {
				logx.Error("Tunnel.serve onTextMessage failed:%v", err.Error())
			}
		} else if messageType == websocket.BinaryMessage {
			if err := t.onBinaryMessage(message); err != nil {
				logx.Error("Tunnel.serve onMessage failed:", err)
			}
		} else {
			logx.Error("unsupport websocket message type:%d", messageType)
		}
	}
}

func (t *Tunnel) onTextMessage(data []byte) error {
	response := &ProxyResponse{}
	err := json.Unmarshal(data, response)
	if err != nil {
		return err
	}

	if len(response.ID) == 0 {
		return fmt.Errorf("invalid response %v", *response)
	}

	headerBuilder := response.rebuildHTTPHeaders()

	contentLength, err := getContentLength(response.Header)
	if err != nil {
		return err
	}

	v, ok := t.reqs.Load(response.ID)
	if !ok {
		return fmt.Errorf("http request %s not exist", response.ID)
	}

	req := v.(*Req)
	req.rspContentLength = contentLength
	return req.write([]byte(headerBuilder.String()))
}

func (t *Tunnel) onBinaryMessage(data []byte) error {
	if len(data) <= uuidLength {
		return fmt.Errorf("invalid data, data length < %d", uuidLength)
	}

	uid, err := uuid.ParseBytes(data[:uuidLength])
	if err != nil {
		return err
	}

	id := uid.String()

	v, ok := t.reqs.Load(id)
	if !ok {
		return fmt.Errorf("can not find request %s", id)
	}

	req := v.(*Req)
	req.rspDataLength = req.rspDataLength + int64(len(data[uuidLength:]))
	req.write(data[uuidLength:])

	if req.rspDataLength >= req.rspContentLength {
		t.reqs.Delete(id)
	}
	return nil
}

func (t *Tunnel) onHTTPRequest(targetInfo *TargetInfo) error {
	req := &Req{ID: targetInfo.req.ID, conn: targetInfo.conn, tun: t}

	t.reqs.Store(req.ID, req)
	defer t.reqs.Delete(req.ID)

	reqData, err := json.Marshal(targetInfo.req)
	if err != nil {
		return err
	}

	logx.Debugf("onHTTPRequest:%v", *targetInfo.req)

	if err := t.writeText(reqData); err != nil {
		return err
	}

	if len(targetInfo.extraBytes) > 0 {
		if err := t.onHTTPData(targetInfo.req.ID, targetInfo.extraBytes); err != nil {
			return err
		}
	}

	return req.proxy()
}

func (t *Tunnel) onHTTPData(id string, data []byte) error {
	uid := uuid.MustParse(id)
	buf := append(uid[:], data...)
	return t.writeBinary(buf)
}

func (t *Tunnel) keepalive() {
	if t.waitping > 3 {
		t.conn.Close()
		return
	}

	b := make([]byte, 8)
	now := time.Now().Unix()
	binary.LittleEndian.PutUint64(b, uint64(now))

	t.writePing(b)

	t.waitping++
}

func (t *Tunnel) writeBinary(msg []byte) error {
	t.writeLock.Lock()
	defer t.writeLock.Unlock()

	return t.conn.WriteMessage(websocket.BinaryMessage, msg)
}

func (t *Tunnel) writeText(msg []byte) error {
	t.writeLock.Lock()
	defer t.writeLock.Unlock()

	return t.conn.WriteMessage(websocket.TextMessage, msg)
}

func (t *Tunnel) close() {
	if t.conn != nil {
		t.conn.Close()
	}

}
