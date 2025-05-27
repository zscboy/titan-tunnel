package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateVMLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateVMLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateVMLogic {
	return &CreateVMLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateVMLogic) CreateVM(req *types.CreateVMRequest) (resp *types.VMOperationResponse, err error) {
	logx.Debugf("CreateVMLogic.CreateVM request %#v", *req)
	request := &vms.CreateVMRequest{
		Id:       req.Id,
		VmName:   req.VmName,
		Cpu:      req.Cpu,
		Memory:   req.Memory,
		DiskSize: req.DiskSize,
		Image:    req.Image,
	}

	rsp, err := l.svcCtx.Vms.CreateVM(l.ctx, request)
	if err != nil {
		return nil, err
	}

	return &types.VMOperationResponse{Success: rsp.Success}, nil
}
