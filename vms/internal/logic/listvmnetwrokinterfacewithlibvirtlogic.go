package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVMNetwrokInterfaceWithLibvirtLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListVMNetwrokInterfaceWithLibvirtLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVMNetwrokInterfaceWithLibvirtLogic {
	return &ListVMNetwrokInterfaceWithLibvirtLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListVMNetwrokInterfaceWithLibvirtLogic) ListVMNetwrokInterfaceWithLibvirt(in *pb.ListVMNetwrokInterfaceReqeust) (*pb.ListVMNetworkInterfaceResponse, error) {
	opts, err := getVirtOpts(l.svcCtx.Redis, in.Id)
	if err != nil {
		return nil, err
	}

	vmAPI := l.svcCtx.Virt.GetVMAPI(opts)
	if vmAPI == nil {
		return nil, fmt.Errorf("can not find vm api:%s", opts.VMAPI)
	}

	return vmAPI.ListVMNetwrokInterfaceWithLibvirt(l.ctx, in)
}
