package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/logx"
)

type StopVMLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStopVMLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StopVMLogic {
	return &StopVMLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StopVMLogic) StopVM(req *types.StopVMRequest) (resp *types.VMOperationResponse, err error) {
	request := &vms.StopVMRequest{
		Id:     req.Id,
		VmName: req.VmName,
	}
	rsp, err := l.svcCtx.Vms.StopVM(l.ctx, request)
	if err != nil {
		return nil, err
	}

	return &types.VMOperationResponse{Success: rsp.GetSuccess(), Message: rsp.GetMessage()}, nil
}
