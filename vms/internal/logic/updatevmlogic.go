package logic

import (
	"context"
	"fmt"

	"titan-vm/vms/internal/svc"
	"titan-vm/vms/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateVMLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateVMLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateVMLogic {
	return &UpdateVMLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateVMLogic) UpdateVM(in *pb.UpdateVMRequest) (*pb.VMOperationResponse, error) {
	opts, err := getVirtOpts(l.svcCtx.Redis, in.Id)
	if err != nil {
		return nil, err
	}

	vmAPI := l.svcCtx.Virt.GetVMAPI(opts)
	if vmAPI == nil {
		return nil, fmt.Errorf("can not find vm api:%s", opts.VMAPI)
	}

	err = vmAPI.UpdateVM(l.ctx, in)
	if err != nil {
		return nil, err
	}
	return &pb.VMOperationResponse{}, nil
}
