package logic

import (
	"context"
	"fmt"

	"titan-tunnel/server/internal/svc"
	"titan-tunnel/server/internal/types"
	"titan-tunnel/server/model"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	actionStart = "start"
	ActonStop   = "stop"
)

type StartOrStopUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStartOrStopUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartOrStopUserLogic {
	return &StartOrStopUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StartOrStopUserLogic) StartOrStopUser(req *types.StartOrStopUserReq) (resp *types.UserOperationResp, err error) {
	if req.Action != actionStart && req.Action != ActonStop {
		return &types.UserOperationResp{ErrMsg: "Actoin not start or stop"}, nil
	}

	user, err := model.GetUser(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if user == nil {
		return &types.UserOperationResp{ErrMsg: fmt.Sprintf("user %s not exist", req.UserName)}, nil
	}

	user.Off = (req.Action == ActonStop)

	err = model.SaveUser(l.svcCtx.Redis, user)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	l.svcCtx.TunMgr.DeleteUserFromCache(req.UserName)
	return &types.UserOperationResp{Success: true}, nil

}
