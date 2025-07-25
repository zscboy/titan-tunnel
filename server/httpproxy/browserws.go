package httpproxy

import (
	"net/http"
	"time"
	"titan-tunnel/server/api/model"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

const (
	timeLayout = "2006-01-02 15:04:05.999999999 -0700 MST"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development.
		},
	} // use default options
)

type WebWSReq struct {
	NodeId string `form:"id,optional"`
	// OS     string `form:"os"`
}

type BrowserWS struct {
	tunMgr *TunnelManager
}

func newBrowserWS(tunMgr *TunnelManager) *BrowserWS {
	return &BrowserWS{tunMgr: tunMgr}
}

func (ws *BrowserWS) ServeWS(w http.ResponseWriter, r *http.Request) error {
	logx.Infof("ServeWS")
	ip, err := getRemoteIP(r)
	if err != nil {
		return err
	}

	var req WebWSReq
	if err := httpx.Parse(r, &req); err != nil {
		return err
	}

	if len(req.NodeId) == 0 {
		req.NodeId = uuid.NewString()
	}

	logx.Infof("WebWS.ServeWS %s, ip: %s", r.URL.Path, ip)

	node, err := model.GetNode(ws.tunMgr.redis, req.NodeId)
	if err != nil {
		logx.Errorf("ServeWS, get node %s", err.Error())
		return err
	}

	if node == nil {
		node = &model.Node{Id: req.NodeId, RegisterAt: time.Now().Format(timeLayout)}
	}

	// browser.OS = req.OS
	node.IP = ip
	node.Online = true
	node.LoginAt = time.Now().Format(timeLayout)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	ws.tunMgr.acceptWebsocket(c, node)

	return nil
}
