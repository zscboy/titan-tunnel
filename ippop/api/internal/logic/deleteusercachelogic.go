package logic

import (
	"context"

	"titan-ipoverlay/ippop/api/internal/svc"
	"titan-ipoverlay/ippop/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserCacheLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteUserCacheLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserCacheLogic {
	return &DeleteUserCacheLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteUserCacheLogic) DeleteUserCache(req *types.DeleteUserCache) error {
	// todo: add your logic here and delete this line
	l.svcCtx.TunMgr.DeleteUserFromCache(req.UserName)
	return nil
}
