package ws

import (
	"context"
	"sync"
	"time"
	pb "titan-vm/pb"
	"titan-vm/vms/internal/config"
	"titan-vm/vms/model"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type TunnelManager struct {
	tunnels sync.Map
	// svcCtx  *svc.ServiceContext
	redis  *redis.Redis
	config config.Config
}

func NewTunnelManager(config config.Config, redis *redis.Redis) *TunnelManager {
	tm := &TunnelManager{config: config, redis: redis}
	go tm.keepalive()
	return tm
}

func (tm *TunnelManager) acceptWebsocket(conn *websocket.Conn, node *model.Node) {
	logx.Debugf("TunnelManager:%s accept websocket ", node.Id)
	// TODO: wait for old tunnel to disconnect
	v, ok := tm.tunnels.Load(node.Id)
	if ok {
		oldTun := v.(*CtrlTunnel)
		oldTun.close()
	}

	ctrlTun := newCtrlTunnel(conn, tm, &TunOptions{Id: node.Id, OS: node.OS, VMAPI: node.VmAPI, IP: node.IP})
	tm.tunnels.Store(node.Id, ctrlTun)

	if err := model.SetNode(context.Background(), tm.redis, node); err != nil {
		logx.Errorf("SetNode failed:%s", err.Error())
		return
	}

	if err := model.SetNodeOnline(tm.redis, node.Id); err != nil {
		logx.Errorf("SetNodeOnline failed:%s", err.Error())
		return
	}

	go func() {
		if err := ctrlTun.authRequest(context.Background()); err != nil {
			logx.Errorf("auth request failed:%v", err)
		} else {
			logx.Debugf("auth %s success", node.Id)
		}

	}()

	defer model.SetNodeOffline(tm.redis, node.Id)
	defer tm.tunnels.Delete(node.Id)

	ctrlTun.serve()
}

// func (tm *TunnelManager) setNodeOffline(node *model.Node) {
// 	node.OfflineAt = time.Now().String()
// 	node.Online = false
// 	model.SaveNode(tm.redis, node)
// }

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
