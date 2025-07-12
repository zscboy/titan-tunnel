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

type SwitchUserRouteNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSwitchUserRouteNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SwitchUserRouteNodeLogic {
	return &SwitchUserRouteNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SwitchUserRouteNodeLogic) SwitchUserRouteNode(req *types.SwitchUserRouteNodeReq) (resp *types.UserOperationResp, err error) {
	user, err := model.GetUser(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	api := l.svcCtx.ServerAPIs[user.PopID]
	if api == nil {
		return &types.UserOperationResp{ErrMsg: fmt.Sprintf("pop %s not found", user.PopID)}, nil
	}

	in := &serverapi.SwitchUserRouteNodeReq{UserName: user.UserName, NodeId: req.NodeId}
	startOrStopResp, err := api.SwitchUserRouteNode(l.ctx, in)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	return &types.UserOperationResp{Success: startOrStopResp.Success, ErrMsg: startOrStopResp.ErrMsg}, nil
}
