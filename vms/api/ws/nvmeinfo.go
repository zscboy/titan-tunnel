package ws

import (
	"context"
	"fmt"
	"titan-vm/pb"
	"titan-vm/vms/api/internal/types"
)

type NvmeInfo struct {
	tunMgr *TunnelManager
}

func NewNvmeInfo(tunMgr *TunnelManager) *NvmeInfo {
	return &NvmeInfo{tunMgr: tunMgr}

}

func (nvmeInfo *NvmeInfo) List(ctx context.Context, req *types.ListNvmeRequest) (*types.ListNvmeResponse, error) {
	tun := nvmeInfo.tunMgr.getTunnel(req.NodeId)
	if tun == nil {
		return nil, fmt.Errorf("not found %s", req.NodeId)
	}

	resp := &pb.CmdListNvmeResponse{}
	payload := &pb.Command{Type: pb.CommandType_LIST_NVME, Data: []byte{}}
	err := tun.sendCommand(ctx, payload, resp)
	if err != nil {
		return nil, err
	}

	nvmes := make([]*types.NvmeInfo, 0, len(resp.Nvmes))
	for _, nvme := range resp.Nvmes {
		nvmeInfo := &types.NvmeInfo{Name: nvme.Name, PciAddr: nvme.PciAddr}
		nvmes = append(nvmes, nvmeInfo)
	}
	return &types.ListNvmeResponse{Nvmes: nvmes}, nil
}
