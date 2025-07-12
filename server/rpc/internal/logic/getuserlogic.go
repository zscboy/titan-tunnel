package logic

import (
	"context"
	"fmt"

	"titan-tunnel/server/api/model"
	"titan-tunnel/server/rpc/internal/svc"
	"titan-tunnel/server/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserLogic) GetUser(in *pb.GetUserReq) (*pb.GetUserResp, error) {
	user, err := model.GetUser(l.svcCtx.Redis, in.UserName)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user %s not exist", in.UserName)
	}

	node, err := model.GetNode(l.svcCtx.Redis, user.RouteNodeID)
	if err != nil {
		return nil, err
	}

	trafficLimit := &pb.TrafficLimit{StartTime: user.StartTime, EndTime: user.EndTime, TotalTraffic: user.TotalTraffic}
	route := &pb.Route{Mode: int32(user.RouteMode), NodeId: user.RouteNodeID, Intervals: int32(user.UpdateRouteIntervals)}
	return &pb.GetUserResp{UserName: in.UserName, TrafficLimit: trafficLimit, Route: route, NodeIp: node.IP}, nil
}
