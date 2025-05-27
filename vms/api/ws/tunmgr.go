package ws

import (
	"context"
	"sync"
	"time"
	pb "titan-vm/pb"
	"titan-vm/vms/model"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type TunnelManager struct {
	tunnels sync.Map
	// svcCtx  *svc.ServiceContext
	redis *redis.Redis
}

func NewTunnelManager(redis *redis.Redis) *TunnelManager {
	tm := &TunnelManager{redis: redis}
	go tm.keepalive()
	return tm
}

func (tm *TunnelManager) acceptWebsocket(conn *websocket.Conn, uuid string, opts *TunOptions) {
	logx.Debugf("TunnelManager:%s accept websocket ", uuid)
	// TODO: wait for old tunnel to disconnect
	v, ok := tm.tunnels.Load(uuid)
	if ok {
		oldTun := v.(*CtrlTunnel)
		oldTun.close()
	}

	ctrlTun := newCtrlTunnel(uuid, conn, opts)
	tm.tunnels.Store(uuid, ctrlTun)

	node, err := model.GetNode(tm.redis, uuid)
	if err != nil {
		logx.Errorf("TunnelManager.acceptWebsocket, get node %s", err.Error())
		return
	}

	if node == nil {
		node = &model.Node{Id: uuid, OS: opts.OS, VmAPI: opts.VMAPI, IP: opts.IP, RegisterAt: time.Now().String()}
		tm.registerNode(node)
	} else {
		node.IP = opts.IP
		tm.setNodeOnline(node)
	}

	defer tm.setNodeOffline(node)
	defer tm.tunnels.Delete(uuid)

	ctrlTun.serve()
}

func (tm *TunnelManager) registerNode(node *model.Node) {
	node.LoginAt = time.Now().String()
	node.Online = true
	model.RegisterNode(context.Background(), tm.redis, node)
}

func (tm *TunnelManager) setNodeOnline(node *model.Node) {
	node.LoginAt = time.Now().String()
	node.Online = true
	model.SaveNode(tm.redis, node)
}

func (tm *TunnelManager) setNodeOffline(node *model.Node) {
	node.OfflineAt = time.Now().String()
	node.Online = false
	model.SaveNode(tm.redis, node)
}

func (tm *TunnelManager) onVmClient(conn *websocket.Conn, uuid string, address *pb.DestAddr, transportType TransportType) {
	logx.Debugf("TunnelManager.onVmClient uuid:%s", uuid)
	v, ok := tm.tunnels.Load(uuid)
	if !ok {
		logx.Errorf("TunnelManager.onVmClient, client %s not exist", uuid)
		return
	}

	tun := v.(*CtrlTunnel)

	if err := tun.onVmClientAcceptRequest(conn, address, transportType); err != nil {
		logx.Errorf("onVmClient, onLibvirtClientAcceptRequest error %s", err.Error())
	}

	logx.Debugf("TunnelManager.onVmClient uuid:%s exit", uuid)
}

func (tm *TunnelManager) keepalive() {
	tick := 0
	for {
		time.Sleep(time.Second * 1)
		tick++

		if tick == 30 {
			tick = 0
			tm.tunnels.Range(func(key, value any) bool {
				t := value.(*CtrlTunnel)
				t.keepalive()
				return true
			})
		}
	}
}

func (tm *TunnelManager) getTunnel(id string) *CtrlTunnel {
	v, ok := tm.tunnels.Load(id)
	if !ok {
		return nil
	}
	return v.(*CtrlTunnel)
}

// func (tm *TunnelManager) waitTunClose(tun *CtrlTunnel) {

// 	tun.close()
// }

// func (tm *TunnelManager) unWaitTunClose(tun *CtrlTunnel) {

// }
