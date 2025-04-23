package server

import (
	"time"
	"titan-vm/proto"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	libvirtLocalSocketAddress = "/var/run/libvirt/libvirt-sock"
)

type TunnelManager struct {
	tunnels map[string]*CtrlTunnel
	// tidx    int
	// maxTunnel uint
}

func newTunnelManager() *TunnelManager {
	return &TunnelManager{tunnels: make(map[string]*CtrlTunnel)}
}

func (tm *TunnelManager) acceptWebsocket(conn *websocket.Conn, uuid string) {
	log.Printf("TunnelManager:%s accept websocket, total:%d", uuid, 1+len(tm.tunnels))
	// TODO: wait for old tunnel to disconnect
	oldTun, ok := tm.tunnels[uuid]
	if ok {
		oldTun.onClose()
	}

	ctrlTun := newCtrlTunnel(uuid, conn)
	tm.tunnels[uuid] = ctrlTun
	defer delete(tm.tunnels, uuid)

	ctrlTun.serve()
}

func (tm *TunnelManager) onLibvirtClient(conn *websocket.Conn, uuid string, targetAddress string) {
	log.Printf("TunnelManager.onLibvirtClient uuid:%s", uuid)
	tun, ok := tm.tunnels[uuid]
	if !ok {
		log.Errorf("TunnelManager.onLibvirtClient, client %s not exist", uuid)
		return
	}

	address := proto.DestAddr{Addr: libvirtLocalSocketAddress, Network: "unix"}
	if len(targetAddress) > 0 {
		address.Addr = targetAddress
		address.Network = "tcp"
	}

	if err := tun.onLibvirtClientAcceptRequest(conn, &address); err != nil {
		log.Errorf("onLibvirtClient, onLibvirtClientAcceptRequest error %s", err.Error())
	}
}

func (tm *TunnelManager) onWebNnc(conn *websocket.Conn, uuid string, targetAddress string) {
	log.Printf("TunnelManager.onWebNnc uuid:%s", uuid)
	tun, ok := tm.tunnels[uuid]
	if !ok {
		log.Errorf("TunnelManager.onWebNnc, client %s not exist", uuid)
		return
	}

	address := proto.DestAddr{Addr: targetAddress, Network: "tcp"}
	tun.onWebVncAcceptRequest(conn, &address)
}

func (tm *TunnelManager) keepalive() {
	tick := 0
	for {
		time.Sleep(time.Second * 1)
		tick++

		if tick == 30 {
			tick = 0
			for _, t := range tm.tunnels {
				t.keepalive()
			}
		}
	}
}
