package logic

import (
	"context"

	"titan-tunnel/server/api/model"
	"titan-tunnel/server/rpc/internal/svc"
	"titan-tunnel/server/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserLogic {
	return &ListUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUserLogic) ListUser(in *pb.ListUserReq) (*pb.ListUserResp, error) {
	users, err := model.ListUser(l.ctx, l.svcCtx.Redis, int(in.Start), int(in.End))
	if err != nil {
		return nil, err
	}

	us := make([]*pb.User, 0, len(users))
	for _, user := range users {
		trafficLimit := pb.TrafficLimit{StartTime: user.StartTime, EndTime: user.EndTime, TotalTraffic: user.TotalTraffic}
		route := pb.Route{Mode: int32(user.RouteMode), NodeId: user.RouteNodeID, Intervals: int32(user.UpdateRouteIntervals)}
		u := &pb.User{UserName: user.UserName, TrafficLimit: &trafficLimit, Route: &route, CurrentTraffic: user.CurrentTraffic, Off: user.Off}
		us = append(us, u)
	}

	// TODO: use redis transactions to get all node
	for _, user := range us {
		if user.Route == nil {
			continue
		}
		route := user.Route

		node, err := model.GetNode(l.svcCtx.Redis, route.NodeId)
		if err != nil {
			logx.Errorf("get node %v", err)
			continue
		}

		user.NodeIp = node.IP
		user.NodeOnline = node.Online
	}

	total, err := model.GetUserLen(l.svcCtx.Redis)
	if err != nil {
		return nil, err
	}

	return &pb.ListUserResp{Users: us, Total: int32(total)}, nil
}
