package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDiskWithLibvirtLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDiskWithLibvirtLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDiskWithLibvirtLogic {
	return &DeleteDiskWithLibvirtLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteDiskWithLibvirtLogic) DeleteDiskWithLibvirt(in *pb.DeleteDiskRequest) (*pb.VMOperationResponse, error) {
	opts, err := getVirtOpts(l.svcCtx.Redis, in.Id)
	if err != nil {
		return nil, err
	}

	vmAPI := l.svcCtx.Virt.GetVMAPI(opts)
	if vmAPI == nil {
		return nil, fmt.Errorf("can not find vm api:%s", opts.VMAPI)
	}

	err = vmAPI.DeleteDiskWithLibvirt(l.ctx, in)
	if err != nil {
		return nil, err
	}

	return &pb.VMOperationResponse{Success: true}, nil
}
