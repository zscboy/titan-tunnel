package ws

import (
	"fmt"
	"net"
	"net/http"
	"titan-vm/pb"
	"titan-vm/vms/api/internal/types"
)

const (
	transportTypeRaw            = "raw"
	transportTypeWebsocket      = "websoket"
	vmapiMultipass              = "multipass"
	vmapiLibvirt                = "libvirt"
	unixSocketFilePathLibvirt   = "/var/run/libvirt/libvirt-sock"
	unixSocketFilePathMultipass = "/var/snap/multipass/common/multipass_socket"
)

type VMWS struct {
	tunMgr *TunnelManager
}

func NewVMWS(tunMgr *TunnelManager) *VMWS {
	return &VMWS{tunMgr: tunMgr}
}

func (ws *VMWS) ServeWS(w http.ResponseWriter, r *http.Request, req *types.VMWSRequest) error {
	if len(req.NodeId) == 0 {
		return fmt.Errorf("require NodeId")
	}

	if len(req.Address) > 0 {
		_, _, err := net.SplitHostPort(req.Address)
		if err != nil {
			return fmt.Errorf("parse Address failed:%s", err.Error())
		}
	}

	if len(req.Address) == 0 && len(req.VMAPI) == 0 {
		return fmt.Errorf("require Address or VMAPI")
	}

	destAddr := &pb.DestAddr{Network: "tcp", Addr: req.Address}
	if len(destAddr.Addr) == 0 {
		socket, err := ws.getUnixSocketFilePath(req.VMAPI)
		if err != nil {
			return err
		}
		destAddr.Addr = socket
		destAddr.Network = "unix"
	}

	var transportType TransportType
	if req.Transport == transportTypeRaw {
		transportType = TransportTypeTcp
	} else if req.Transport == transportTypeWebsocket {
		transportType = TransportTypeWebsocket
	} else {
		return fmt.Errorf("unsupport transport type %s", req.Transport)
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	ws.tunMgr.onVmClient(c, req.NodeId, destAddr, transportType)

	return nil
}

func (ws *VMWS) getUnixSocketFilePath(vmapi string) (string, error) {
	switch vmapi {
	case vmapiMultipass:
		return unixSocketFilePathMultipass, nil
	case vmapiLibvirt:
		return unixSocketFilePathLibvirt, nil
	default:
		return "", fmt.Errorf("unsupport vmapi %s", vmapi)
	}
}
