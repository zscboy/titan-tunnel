package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVMInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetVMInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVMInfoLogic {
	return &GetVMInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVMInfoLogic) GetVMInfo(req *types.VMInfoReqeust) (resp *types.VMInfoResponse, err error) {
	request := &vms.GetVMInfoRequest{
		Id:     req.Id,
		VmName: req.VmName,
	}
	rsp, err := l.svcCtx.Vms.GetVMInfo(l.ctx, request)
	if err != nil {
		return nil, err
	}

	disks := make([]*types.VMDisk, 0, len(rsp.Disks))
	for _, d := range rsp.Disks {
		disk := &types.VMDisk{
			DiskType:   uint32(d.DiskType),
			SourcePath: d.SourcePath,
			TargetDev:  d.TargetDev,
			TargetBus:  d.TargetBus,
		}

		disks = append(disks, disk)
	}

	interfaces := make([]*types.VMNetworkInterface, 0, len(rsp.Interfaces))
	for _, iface := range rsp.Interfaces {
		itface := &types.VMNetworkInterface{
			Name:        iface.Name,
			Type:        iface.Type,
			Source:      iface.Source,
			SourceModel: iface.SourceModel,
			Model:       iface.Model,
			Mac:         iface.Mac,
		}
		interfaces = append(interfaces, itface)
	}

	hostdevs := make([]*types.VMHostdev, 0, len(rsp.Hostdevs))
	for _, d := range rsp.Hostdevs {
		hostdev := &types.VMHostdev{
			SourceAddrDomain: uint32(d.SourceAddrDomain),
			SourceAddrBus:    uint32(d.SourceAddrBus),
			SourceAddrSlot:   uint32(d.SourceAddrSlot),
		}
		hostdevs = append(hostdevs, hostdev)
	}
	return &types.VMInfoResponse{CPU: rsp.Cpu, Memroy: rsp.Memory, Disks: disks, Interfaces: interfaces, Hostdevs: hostdevs, VncPort: rsp.VncPort}, nil
}
