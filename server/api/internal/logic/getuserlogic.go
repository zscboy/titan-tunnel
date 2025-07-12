package logic

import (
	"context"
	"fmt"

	"titan-tunnel/server/api/internal/svc"
	"titan-tunnel/server/api/internal/types"
	"titan-tunnel/server/api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserLogic) GetUser(req *types.GetUserReq) (resp *types.GetUserResp, err error) {
	user, err := model.GetUser(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user %s not exist", req.UserName)
	}

	node, err := model.GetNode(l.svcCtx.Redis, user.RouteNodeID)
	if err != nil {
		return nil, err
	}

	trafficLimit := &types.TrafficLimit{StartTime: user.StartTime, EndTime: user.EndTime, TotalTraffic: user.TotalTraffic}
	route := &types.Route{Mode: user.RouteMode, NodeID: user.RouteNodeID, Intervals: user.UpdateRouteIntervals}
	return &types.GetUserResp{UserName: req.UserName, TrafficLimit: trafficLimit, Route: route, NodeIP: node.IP}, nil
}
