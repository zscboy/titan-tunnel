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

type ModifyUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewModifyUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyUserLogic {
	return &ModifyUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyUserLogic) ModifyUser(req *types.ModifyUserReq) (resp *types.UserOperationResp, err error) {
	popID, err := model.GetUserPop(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if len(popID) == 0 {
		return &types.UserOperationResp{ErrMsg: fmt.Sprintf("user %s not exist", req.UserName)}, nil
	}

	server := l.svcCtx.Servers[popID]
	if server == nil {
		return &types.UserOperationResp{ErrMsg: fmt.Sprintf("pop %s not found", popID)}, nil
	}

	in := &serverapi.ModifyUserReq{UserName: req.UserName}
	if req.TrafficLimit != nil {
		in.TrafficLimit = toTrafficLimitReq(req.TrafficLimit)
	}

	if req.Route != nil {
		in.Route = toRouteReq(req.Route)
	}

	modifyUserResp, err := server.API.ModifyUser(l.ctx, in)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	return &types.UserOperationResp{Success: modifyUserResp.Success, ErrMsg: modifyUserResp.ErrMsg}, nil
}
