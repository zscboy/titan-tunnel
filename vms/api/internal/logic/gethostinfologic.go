package logic

import (
	"context"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/api/ws"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHostInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetHostInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHostInfoLogic {
	return &GetHostInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHostInfoLogic) GetHostInfo(req *types.HostInfoRequest) (resp *types.HostInfoResponse, err error) {
	hostInfo := ws.NewHostInfo(l.svcCtx.TunMgr)
	return hostInfo.Get(l.ctx, req)
}
