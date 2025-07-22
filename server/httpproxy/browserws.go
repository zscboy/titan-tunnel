package httpproxy

import (
	"net"
	"net/http"
	"strings"
	"time"
	"titan-tunnel/server/api/model"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	timeLayout = "2006-01-02 15:04:05.999999999 -0700 MST"
)

var (
	upgrader = websocket.Upgrader{} // use default options
)

type WebWSReq struct {
	NodeId string `form:"id"`
	OS     string `form:"os"`
}

type BrowserWS struct {
	tunMgr *TunnelManager
}

func newBrowserWS(tunMgr *TunnelManager) *BrowserWS {
	return &BrowserWS{tunMgr: tunMgr}
}

func (ws *BrowserWS) ServeWS(w http.ResponseWriter, r *http.Request, req *WebWSReq) error {
	logx.Infof("WebWS.ServeWS %s, %v", r.URL.Path, req)

	ip, err := ws.getRemoteIP(r)
	if err != nil {
		return err
	}

	browser, err := model.GetBrowser(ws.tunMgr.redis, req.NodeId)
	if err != nil {
		logx.Errorf("ServeWS, get node %s", err.Error())
		return err
	}

	if browser == nil {
		browser = &model.Browser{Id: req.NodeId, RegisterAt: time.Now().Format(timeLayout)}
	}

	browser.OS = req.OS
	browser.IP = ip
	browser.Online = true
	browser.LoginAt = time.Now().Format(timeLayout)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	ws.tunMgr.acceptWebsocket(c, browser)

	return nil
}

func (ws *BrowserWS) getRemoteIP(r *http.Request) (string, error) {
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
