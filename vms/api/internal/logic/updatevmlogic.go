package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateVMLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateVMLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateVMLogic {
	return &UpdateVMLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateVMLogic) UpdateVM(req *types.UpdateVMRequest) (resp *types.VMOperationResponse, err error) {
	request := &vms.UpdateVMRequest{
		Id:     req.Id,
		VmName: req.VmName,
	}
	rsp, err := l.svcCtx.Vms.UpdateVM(l.ctx, request)
	if err != nil {
		return nil, err
	}

	return &types.VMOperationResponse{Success: rsp.GetSuccess(), Message: rsp.GetMessage()}, nil

	return
}
