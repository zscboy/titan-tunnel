package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateVMWithLibvirtLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateVMWithLibvirtLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateVMWithLibvirtLogic {
	return &CreateVMWithLibvirtLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Libvirt 相关操作
func (l *CreateVMWithLibvirtLogic) CreateVMWithLibvirt(in *pb.CreateVMWithLibvirtRequest) (*pb.VMOperationResponse, error) {
	opts, err := getVirtOpts(l.svcCtx.Redis, in.Id)
	if err != nil {
		return nil, err
	}

	vmAPI := l.svcCtx.Virt.GetVMAPI(opts)
	if vmAPI == nil {
		return nil, fmt.Errorf("can not find vm api:%s", opts.VMAPI)
	}

	err = vmAPI.CreateVMWithLibvirt(l.ctx, in)
	if err != nil {
		return nil, err
	}

	return &pb.VMOperationResponse{Success: true}, nil
}
