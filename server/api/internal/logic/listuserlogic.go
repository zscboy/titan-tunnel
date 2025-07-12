package logic

import (
	"context"

	"titan-tunnel/server/api/internal/svc"
	"titan-tunnel/server/api/internal/types"
	"titan-tunnel/server/api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserLogic {
	return &ListUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserLogic) ListUser(req *types.ListUserReq) (resp *types.ListUserResp, err error) {
	users, err := model.ListUser(l.ctx, l.svcCtx.Redis, req.Start, req.End)
	if err != nil {
		return nil, err
	}

	us := make([]*types.User, 0, len(users))
	for _, user := range users {
		trafficLimit := types.TrafficLimit{StartTime: user.StartTime, EndTime: user.EndTime, TotalTraffic: user.TotalTraffic}
		route := types.Route{Mode: user.RouteMode, NodeID: user.RouteNodeID, Intervals: user.UpdateRouteIntervals}
		u := &types.User{UserName: user.UserName, TrafficLimit: &trafficLimit, Route: &route, CurrentTraffic: user.CurrentTraffic, Off: user.Off}
		us = append(us, u)
	}

	total, err := model.GetUserLen(l.svcCtx.Redis)
	if err != nil {
		return nil, err
	}

	return &types.ListUserResp{Users: us, Total: total}, nil
}
