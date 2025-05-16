package ws

import (
	"fmt"
	"net"
	"net/http"
	pb "titan-vm/pb"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	transportTypeRaw            = "raw"
	transportTypeWebsocket      = "websoket"
	vmapiMultipass              = "multipass"
	vmapiLibvirt                = "libvirt"
	unixSocketFilePathLibvirt   = "/var/run/libvirt/libvirt-sock"
	unixSocketFilePathMultipass = "/var/snap/multipass/common/multipass_socket"
)

var (
	// upgrader = websocket.Upgrader{} // use default options
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type WsHandler struct {
	tunMgr *TunnelManager
}

func newWsHandler(tunMgr *TunnelManager) *WsHandler {
	return &WsHandler{tunMgr: tunMgr}
}

func (ws *WsHandler) nodeHandler(w http.ResponseWriter, r *http.Request) {
	logx.Debugf("nodeHandler %s", r.URL.Path)
	var uuid = r.URL.Query().Get("uuid")
	if uuid == "" {
		w.WriteHeader(http.StatusBadRequest)
		logx.Error("need uuid!")
		return
	}

	var os = r.URL.Query().Get("os")
	if os == "" {
		w.WriteHeader(http.StatusBadRequest)
		logx.Error("need os!")
		return
	}

	var vmapi = r.URL.Query().Get("vmapi")
	if vmapi != vmapiMultipass && vmapi != vmapiLibvirt {
		w.WriteHeader(http.StatusBadRequest)
		logx.Errorf("unsupport vmapi: %s", vmapi)
		return
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logx.Error("upgrade:", err)
		return
	}
	defer c.Close()

	ws.tunMgr.acceptWebsocket(c, uuid, &TunOptions{OS: os, VMAPI: vmapi})
}

func (ws *WsHandler) vmHandler(w http.ResponseWriter, r *http.Request) {
	var uuid = r.URL.Query().Get("uuid")
	if uuid == "" {
		logx.Error("WebsocketServer.vmHandler need uuid!")
		return
	}

	var address = r.URL.Query().Get("address")
	if len(address) > 0 {
		_, _, err := net.SplitHostPort(address)
		if err != nil {
			logx.Errorf("WebsocketServer.vmHandler can not parse address %s\n", address)
			return
		}
	}

	var vmapi = r.URL.Query().Get("vmapi")
	if len(address) == 0 && len(vmapi) == 0 {
		logx.Errorf("WebsocketServer.vmHandler params need address or vmapi")
		return
	}

	destAddr := &pb.DestAddr{Network: "tcp", Addr: address}
	if len(destAddr.Addr) == 0 {
		socket, err := getUnixSocketFilePath(vmapi)
		if err != nil {
			logx.Error(err)
			return
		}
		destAddr.Addr = socket
		destAddr.Network = "unix"
	}

	var transportType TransportType
	var transportTypeStr = r.URL.Query().Get("transport")
	if transportTypeStr == transportTypeRaw {
		transportType = TransportTypeTcp
	} else if transportTypeStr == transportTypeWebsocket {
		transportType = TransportTypeWebsocket
	} else {
		logx.Error("WebsocketServer.vmHandler unsupport transport type:", transportTypeStr)
		return
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Error("upgrade:", err)
		return
	}
	defer c.Close()

	ws.tunMgr.onVmClient(c, uuid, destAddr, transportType)
}

func (ws *WsHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "Hello, Stupid!")
}

func getUnixSocketFilePath(vmapi string) (string, error) {
	switch vmapi {
	case vmapiMultipass:
		return unixSocketFilePathMultipass, nil
	case vmapiLibvirt:
		return unixSocketFilePathLibvirt, nil
	default:
		return "", fmt.Errorf("unsupport vmapi %s", vmapi)
	}
}
