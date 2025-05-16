package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteVMLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteVMLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteVMLogic {
	return &DeleteVMLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteVMLogic) DeleteVM(req *types.DeleteVMRequest) (resp *types.VMOperationResponse, err error) {
	request := &vms.DeleteVMRequest{
		Id:     req.Id,
		VmName: req.VmName,
	}
	rsp, err := l.svcCtx.Vms.DeleteVM(l.ctx, request)
	if err != nil {
		return nil, err
	}

	return &types.VMOperationResponse{Success: rsp.GetSuccess(), Message: rsp.GetMessage()}, nil
}
