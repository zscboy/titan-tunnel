package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/vms"

	"github.com/zeromicro/go-zero/core/logx"
)

type MultipassExecLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMultipassExecLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MultipassExecLogic {
	return &MultipassExecLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MultipassExecLogic) MultipassExec(req *types.MultipassExecRequest) (resp *types.MultipassExecResponse, err error) {
	request := &vms.MultipassExecRequest{
		Id:           req.Id,
		InstanceName: req.InstanceName,
		Command:      req.Command,
	}
	rsp, err := l.svcCtx.Vms.MultipassExec(l.ctx, request)
	if err != nil {
		return nil, err
	}

	return &types.MultipassExecResponse{Output: rsp.GetOutput()}, nil
}
