package ws

import (
	"fmt"
	"net"
	"net/http"
	"titan-vm/vms/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type NodeWS struct {
	tunMgr *TunnelManager
}

func NewNodeWS(tunMgr *TunnelManager) *NodeWS {
	return &NodeWS{tunMgr: tunMgr}
}

func (ws *NodeWS) ServeWS(w http.ResponseWriter, r *http.Request, req *types.NodeWSRequest) error {
	logx.Infof("nodeHandler %s", r.URL.Path)
	if len(req.NodeId) == 0 {
		return fmt.Errorf("request NodeId")
	}

	if len(req.OS) == 0 {
		return fmt.Errorf("request OS")
	}

	if len(req.VMAPI) == 0 {
		return fmt.Errorf("request VMAPI")
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return err
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	ws.tunMgr.acceptWebsocket(c, &TunOptions{Id: req.NodeId, OS: req.OS, VMAPI: req.VMAPI, IP: ip})

	return nil
}
