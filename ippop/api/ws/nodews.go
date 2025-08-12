package ws

import (
	"net"
	"net/http"
	"strings"
	"time"
	"titan-ipoverlay/ippop/api/internal/types"
	"titan-ipoverlay/ippop/api/model"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	upgrader = websocket.Upgrader{} // use default options
)

type NodeWS struct {
	tunMgr *TunnelManager
}

func NewNodeWS(tunMgr *TunnelManager) *NodeWS {
	return &NodeWS{tunMgr: tunMgr}
}

func (ws *NodeWS) ServeWS(w http.ResponseWriter, r *http.Request, req *types.NodeWSReq) error {
	logx.Infof("NodeWS.ServeWS %s, %v", r.URL.Path, req)

	ip, err := ws.getRemoteIP(r)
	if err != nil {
		return err
	}

	node, err := model.GetNode(ws.tunMgr.redis, req.NodeId)
	if err != nil {
		logx.Errorf("ServeWS, get node %s", err.Error())
		return err
	}

	if node == nil {
		node = &model.Node{Id: req.NodeId, RegisterAt: time.Now().Format(model.TimeLayout)}
	}

	node.OS = req.OS
	node.IP = ip
	node.Online = true
	node.LoginAt = time.Now().Format(model.TimeLayout)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	ws.tunMgr.acceptWebsocket(c, node)

	return nil
}

func (ws *NodeWS) getRemoteIP(r *http.Request) (string, error) {
	ip := r.Header.Get("X-Real-IP")
	if len(ip) != 0 {
		return ip, nil
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			return ip, nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	return ip, nil
}
