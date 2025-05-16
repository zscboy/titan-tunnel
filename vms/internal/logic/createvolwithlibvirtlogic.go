package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateVolWithLibvirtLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateVolWithLibvirtLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateVolWithLibvirtLogic {
	return &CreateVolWithLibvirtLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateVolWithLibvirtLogic) CreateVolWithLibvirt(in *pb.CreateVolWithLibvirtReqeust) (*pb.CreateVolWithLibvirtResponse, error) {
	opts, err := getVirtOpts(l.svcCtx.Redis, in.Id)
	if err != nil {
		return nil, err
	}

	vmAPI := l.svcCtx.Virt.GetVMAPI(opts)
	if vmAPI == nil {
		return nil, fmt.Errorf("can not find vm api:%s", opts.VMAPI)
	}

	return vmAPI.CreateVolWithLibvirt(l.ctx, in)
}
