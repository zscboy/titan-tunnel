package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"titan-tunnel/server/api/model"
	"titan-tunnel/server/rpc/internal/svc"
	"titan-tunnel/server/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// 1000 GB
	defaultTotalTraffic = 1000
	trafficUnit         = 1024 * 1024 * 1024
)

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateUserLogic) CreateUser(in *pb.CreateUserReq) (*pb.CreateUserResp, error) {
	if in.PopId != l.svcCtx.Config.PopID {
		return nil, fmt.Errorf("pop id not match, require pop id %s", l.svcCtx.Config.PopID)
	}
	user, err := model.GetUser(l.svcCtx.Redis, in.UserName)
	if err != nil {
		return nil, err
	}

	if user != nil {
		return nil, fmt.Errorf("user %s already exist", in.UserName)
	}

	if in.Route != nil {
		if err := checkRoute(l.svcCtx.Redis, in.Route); err != nil {
			return nil, err
		}
	}

	if in.TrafficLimit != nil {
		if err := checkTraffic(in.TrafficLimit); err != nil {
			return nil, err
		}
	}

	hash := md5.Sum([]byte(in.Password))
	passwordMD5 := hex.EncodeToString(hash[:])

	trafficLimit := in.TrafficLimit
	if trafficLimit == nil {
		trafficLimit = l.defaultTrafficLimit()
	}

	route := in.Route
	if route == nil {
		route = l.defaultRoute()
	}

	if len(route.NodeId) == 0 {
		route.NodeId = l.allocateNode()
	}

	if len(route.NodeId) == 0 {
		return nil, fmt.Errorf("no enough node for user")
	}

	user = &model.User{
		UserName:             in.UserName,
		PasswordMD5:          passwordMD5,
		StartTime:            trafficLimit.StartTime,
		EndTime:              trafficLimit.EndTime,
		TotalTraffic:         trafficLimit.TotalTraffic * trafficUnit,
		RouteMode:            int(route.Mode),
		RouteNodeID:          route.NodeId,
		UpdateRouteIntervals: int(route.Intervals),
		UpdateRouteTime:      0,
	}

	err = model.SaveUser(l.svcCtx.Redis, user)
	if err != nil {
		return nil, err
	}

	err = model.ZaddUser(l.svcCtx.Redis, user.UserName)
	if err != nil {
		return nil, err
	}

	err = model.BindNode(l.svcCtx.Redis, route.NodeId, in.UserName)
	if err != nil {
		return nil, err
	}

	node, err := model.GetNode(l.svcCtx.Redis, route.NodeId)
	if err != nil {
		return nil, err
	}

	createUserResp := &pb.CreateUserResp{
		UserName:     in.UserName,
		TrafficLimit: trafficLimit,
		Route:        route,
		NodeIp:       node.IP,
	}

	return createUserResp, nil

}

func (l *CreateUserLogic) defaultRoute() *pb.Route {
	return &pb.Route{Mode: routeModeTypeManual, NodeId: l.allocateNode(), Intervals: 0}
}

func (l *CreateUserLogic) allocateNode() string {
	nodeID, err := model.GetOnlineAndUnbindNode(l.svcCtx.Redis)
	if err != nil {
		logx.Errorf("GetOnlineAndUnbindNode %v", err)
		return ""
	}
	return nodeID
}

func (l *CreateUserLogic) defaultTrafficLimit() *pb.TrafficLimit {
	return &pb.TrafficLimit{
		StartTime:    time.Now().Unix(),
		EndTime:      time.Now().AddDate(0, 1, 0).Unix(),
		TotalTraffic: defaultTotalTraffic,
	}
}
