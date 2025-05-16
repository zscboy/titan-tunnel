package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartVMLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStartVMLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartVMLogic {
	return &StartVMLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StartVMLogic) StartVM(req *types.StartVMRequest) (resp *types.VMOperationResponse, err error) {
	request := &vms.StartVMRequest{
		Id:     req.Id,
		VmName: req.VmName,
	}
	rsp, err := l.svcCtx.Vms.StartVM(l.ctx, request)
	if err != nil {
		return nil, err
	}

	return &types.VMOperationResponse{Success: rsp.GetSuccess(), Message: rsp.GetMessage()}, nil
}
