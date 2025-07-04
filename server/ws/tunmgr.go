package ws

import (
	"fmt"
	"net"
	"sync"
	"time"
	"titan-tunnel/server/internal/config"
	"titan-tunnel/server/model"
	"titan-tunnel/server/socks5"

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
	v, ok := tm.tunnels.Load(node.Id)
	if ok {
		oldTun := v.(*Tunnel)
		oldTun.close()
	}

	socksConfig := tm.config.Socks5
	tun := newTunnel(conn, tm, &TunOptions{Id: node.Id, OS: node.OS, VMAPI: node.VmAPI, IP: node.IP, UDPTimeout: int(socksConfig.UDPTimeout), TCPTimeout: int(socksConfig.TCPTimeout)})

	tm.tunnels.Store(node.Id, tun)
	defer tm.tunnels.Delete(node.Id)

	// if err := model.SetNodeWithZadd(context.Background(), tm.redis, node); err != nil {
	// 	logx.Errorf("SetNode failed:%s", err.Error())
	// 	return
	// }

	// if err := model.SetNodeOnline(tm.redis, node.Id); err != nil {
	// 	logx.Errorf("SetNodeOnline failed:%s", err.Error())
	// 	return
	// }

	// defer model.SetNodeOffline(tm.redis, node.Id)
	// defer tm.tunnels.Delete(node.Id)

	tun.serve()
}

func (tm *TunnelManager) randomTunnel() *Tunnel {
	var tunnel *Tunnel
	tm.tunnels.Range(func(key, value any) bool {
		tunnel = value.(*Tunnel)
		return false
	})

	return tunnel
}

func (tm *TunnelManager) getTunnelByUser(user string) *Tunnel {
	// TODO: get tunnel by user
	return tm.randomTunnel()
}

func (tm *TunnelManager) HandleSocks5TCP(tcpConn *net.TCPConn, targetInfo *socks5.SocksTargetInfo) error {
	logx.Debugf("HandleSocks5TCP, user %s, DomainName %s, port %d", targetInfo.User, targetInfo.DomainName, targetInfo.Port)
	tun := tm.getTunnelByUser(targetInfo.User)
	if tun == nil {
		return fmt.Errorf("can not allocate tunnel, user %s", targetInfo.User)
	}

	return tun.acceptSocks5TCPConn(tcpConn, targetInfo)
}

func (tm *TunnelManager) HandleSocks5UDP(udpConn socks5.UDPConn, udpInfo *socks5.Socks5UDPInfo, data []byte) error {
	tun := tm.getTunnelByUser(udpInfo.User)
	if tun == nil {
		return fmt.Errorf("can not allocate tunnel, user %s", udpInfo.User)
	}

	return tun.acceptSocks5UDPData(udpConn, udpInfo, data)
}

func (tm *TunnelManager) HandleUserAuth(userName, password string) bool {
	logx.Debugf("HandleUserAuth userName %s password %s", userName, password)
	if userName != "admin" || password != "123456" {
		return false
	}
	return true
}

func (tm *TunnelManager) keepalive() {
	tick := 0
	for {
		time.Sleep(time.Second * 1)
		tick++

		if tick == 30 {
			tick = 0
			tm.tunnels.Range(func(key, value any) bool {
				t := value.(*Tunnel)
				t.keepalive()
				return true
			})
		}
	}
}

func (tm *TunnelManager) getTunnel(id string) *Tunnel {
	v, ok := tm.tunnels.Load(id)
	if !ok {
		return nil
	}
	return v.(*Tunnel)
}
