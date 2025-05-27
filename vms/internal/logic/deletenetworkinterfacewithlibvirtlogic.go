package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteNetworkInterfaceWithLibvirtLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteNetworkInterfaceWithLibvirtLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteNetworkInterfaceWithLibvirtLogic {
	return &DeleteNetworkInterfaceWithLibvirtLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteNetworkInterfaceWithLibvirtLogic) DeleteNetworkInterfaceWithLibvirt(in *pb.DeleteNetworkInterfaceRequest) (*pb.VMOperationResponse, error) {
	opts, err := getVirtOpts(l.svcCtx.Redis, in.Id)
	if err != nil {
		return nil, err
	}

	vmAPI := l.svcCtx.Virt.GetVMAPI(opts)
	if vmAPI == nil {
		return nil, fmt.Errorf("can not find vm api:%s", opts.VMAPI)
	}

	err = vmAPI.DeleteNetworkInterfaceWithLibvirt(l.ctx, in)
	if err != nil {
		return nil, err
	}
	return &pb.VMOperationResponse{Success: true}, nil
}
