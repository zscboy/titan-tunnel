package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/api/ws"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadTaskGetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadTaskGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadTaskGetLogic {
	return &DownloadTaskGetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadTaskGetLogic) DownloadTaskGet(req *types.DownloadTaskGetRequest) (resp *types.DownloadTask, err error) {
	task := ws.NewDownloadtask(l.svcCtx.TunMgr)
	return task.Get(l.ctx, req)
}
