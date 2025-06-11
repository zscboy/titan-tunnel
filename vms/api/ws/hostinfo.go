package ws

import (
	"context"
	"fmt"
	"titan-vm/pb"
	"titan-vm/vms/api/internal/types"
)

type HostInfo struct {
	tunMgr *TunnelManager
}

func NewHostInfo(tunMgr *TunnelManager) *HostInfo {
	return &HostInfo{tunMgr: tunMgr}

}

func (hostInfo *HostInfo) Get(ctx context.Context, req *types.HostInfoRequest) (*types.HostInfoResponse, error) {
	tun := hostInfo.tunMgr.getTunnel(req.NodeId)
	if tun == nil {
		return nil, fmt.Errorf("not found %s", req.NodeId)
	}

	resp := &pb.CmdHostInfoResponse{}
	payload := &pb.Command{Type: pb.CommandType_HOST_INFO, Data: []byte{}}
	err := tun.sendCommand(ctx, payload, resp)
	if err != nil {
		return nil, err
	}

	return hostInfo.responeToTypeHostInfo(resp), nil
}

func (hostInfo *HostInfo) responeToTypeHostInfo(response *pb.CmdHostInfoResponse) *types.HostInfoResponse {
	cpu := types.CPU{}
	if response.Cpu != nil {
		cpu.Num = response.Cpu.Num
		cpu.Arch = response.Cpu.Arch
		cpu.Usage = response.Cpu.Usage
		cpu.Vendor = response.Cpu.Vendor
	}

	memory := types.Memory{}
	if response.Memory != nil {
		memory.Total = response.Memory.Total
		memory.Used = response.Memory.Used
	}

	disks := make([]types.Disk, 0, len(response.Disks))
	for _, d := range response.Disks {
		disk := types.Disk{Name: d.Name, Type: d.Type, Total: int64(d.Total), Used: int64(d.Used), MountPoint: d.MountPoint, FsType: d.FsType}
		disks = append(disks, disk)
	}

	interfaces := make([]types.NetworkInterface, 0, len(response.Interfaces))
	for _, iface := range response.Interfaces {
		networkInterface := types.NetworkInterface{Name: iface.Name, Mac: iface.Mac, State: hostInfo.interfaceState(iface.Flags)}
		interfaces = append(interfaces, networkInterface)

	}

	return &types.HostInfoResponse{CPU: cpu, Memory: memory, Disks: disks, Interfaces: interfaces}
}

func (hostInfo *HostInfo) interfaceState(flags uint32) string {
	if flags&uint32(pb.NetworkInterfaceFlags_UP) != 0 {
		return "up"
	} else {
		return "down"
	}
}
