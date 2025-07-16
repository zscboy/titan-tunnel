package logic

import (
	"context"
	"fmt"

	"titan-tunnel/manager/internal/svc"
	"titan-tunnel/manager/internal/types"
	"titan-tunnel/manager/model"
	"titan-tunnel/server/rpc/serverapi"

	"github.com/zeromicro/go-zero/core/logx"
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
	user, err := model.GetUser(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	server := l.svcCtx.Servers[user.PopID]
	if server == nil {
		return &types.UserOperationResp{ErrMsg: fmt.Sprintf("pop %s not found", user.PopID)}, nil
	}

	in := &serverapi.StartOrStopUserReq{UserName: user.UserName, Action: req.Action}
	startOrStopResp, err := server.API.StartOrStopUser(l.ctx, in)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	return &types.UserOperationResp{Success: startOrStopResp.Success, ErrMsg: startOrStopResp.ErrMsg}, nil
}
