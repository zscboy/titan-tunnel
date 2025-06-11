package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/api/ws"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadTaskDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadTaskDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadTaskDeleteLogic {
	return &DownloadTaskDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadTaskDeleteLogic) DownloadTaskDelete(req *types.DownloadTaskDeleteRequest) error {
	task := ws.NewDownloadtask(l.svcCtx.TunMgr)
	return task.Delete(l.ctx, req)
}
