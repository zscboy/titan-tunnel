package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListHostNetworkInterfaceWithLibvirtLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListHostNetworkInterfaceWithLibvirtLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListHostNetworkInterfaceWithLibvirtLogic {
	return &ListHostNetworkInterfaceWithLibvirtLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListHostNetworkInterfaceWithLibvirtLogic) ListHostNetworkInterfaceWithLibvirt(in *pb.ListHostNetworkInterfaceRequest) (*pb.ListHostNetworkInterfaceResponse, error) {
	opts, err := getVirtOpts(l.svcCtx.Redis, in.Id)
	if err != nil {
		return nil, err
	}

	vmAPI := l.svcCtx.Virt.GetVMAPI(opts)
	if vmAPI == nil {
		return nil, fmt.Errorf("can not find vm api:%s", opts.VMAPI)
	}

	return vmAPI.ListHostNetworkInterfaceWithLibvirt(l.ctx, in)
}
