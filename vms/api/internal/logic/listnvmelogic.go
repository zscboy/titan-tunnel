package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/api/ws"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListNvmeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListNvmeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNvmeLogic {
	return &ListNvmeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListNvmeLogic) ListNvme(req *types.ListNvmeRequest) (resp *types.ListNvmeResponse, err error) {
	nvmeInfo := ws.NewNvmeInfo(l.svcCtx.TunMgr)
	return nvmeInfo.List(l.ctx, req)
}
