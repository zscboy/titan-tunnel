package logic

import (
	"context"
	"fmt"

	"titan-tunnel/server/internal/svc"
	"titan-tunnel/server/internal/types"
	"titan-tunnel/server/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteUserLogic) DeleteUser(req *types.DeleteUserReq) (resp *types.UserOperationResp, err error) {
	user, err := model.GetUser(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user %s not exist", req.UserName)
	}

	err = model.DeleteUser(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if len(user.RouteNodeID) > 0 {
		if err := model.UnbindNode(l.svcCtx.Redis, user.RouteNodeID); err != nil {
			return nil, err
		}
	}

	l.svcCtx.TunMgr.DeleteUserFromCache(req.UserName)
	return &types.UserOperationResp{Success: true}, nil
}
