package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVMInstanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListVMInstanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVMInstanceLogic {
	return &ListVMInstanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListVMInstanceLogic) ListVMInstance(req *types.ListVMInstanceReqeust) (resp *types.ListVMInstanceResponse, err error) {
	// todo: add your logic here and delete this line
	request := &vms.ListVMInstanceReqeust{
		Id: req.Id,
	}
	rsp, err := l.svcCtx.Vms.ListVMInstance(l.ctx, request)
	if err != nil {
		return nil, err
	}

	vmInfos := make([]types.VMInfo, 0, len(rsp.GetVmInfos()))
	for _, info := range rsp.GetVmInfos() {
		vmInfo := types.VMInfo{
			Name:  info.Name,
			State: info.State,
			Ip:    info.Ip,
			Image: info.Image,
		}
		vmInfos = append(vmInfos, vmInfo)
	}

	return &types.ListVMInstanceResponse{VmInfos: vmInfos}, nil
}
