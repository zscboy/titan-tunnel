package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVMDiskWithLibvirtLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListVMDiskWithLibvirtLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVMDiskWithLibvirtLogic {
	return &ListVMDiskWithLibvirtLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListVMDiskWithLibvirtLogic) ListVMDiskWithLibvirt(in *pb.ListVMDiskRequest) (*pb.ListVMDiskResponse, error) {
	opts, err := getVirtOpts(l.svcCtx.Redis, in.Id)
	if err != nil {
		return nil, err
	}

	vmAPI := l.svcCtx.Virt.GetVMAPI(opts)
	if vmAPI == nil {
		return nil, fmt.Errorf("can not find vm api:%s", opts.VMAPI)
	}

	return vmAPI.ListVMDiskWithLibvirt(l.ctx, in)
}
