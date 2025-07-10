package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"titan-tunnel/server/internal/svc"
	"titan-tunnel/server/internal/types"
	"titan-tunnel/server/model"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// 1000 GB
	defaultTotalTraffic = 1000
	trafficUnit         = 1024 * 1024 * 1024
)

type CreateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateUserLogic) CreateUser(req *types.CreateUserReq) (resp *types.CreateUserResp, err error) {
	if req.PopId != l.svcCtx.Config.PopID {
		return nil, fmt.Errorf("pop id not match, require pop id %s", l.svcCtx.Config.PopID)
	}
	user, err := model.GetUser(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return nil, err
	}

	if user != nil {
		return nil, fmt.Errorf("user %s already exist", req.UserName)
	}

	if req.Route != nil {
		if err := checkRoute(l.svcCtx.Redis, req.Route); err != nil {
			return nil, err
		}
	}

	if req.TrafficLimit != nil {
		if err := checkTraffic(req.TrafficLimit); err != nil {
			return nil, err
		}
	}

	hash := md5.Sum([]byte(req.Password))
	passwordMD5 := hex.EncodeToString(hash[:])

	trafficLimit := req.TrafficLimit
	if trafficLimit == nil {
		trafficLimit = l.defaultTrafficLimit()
	}

	route := req.Route
	if route == nil {
		route = l.defaultRoute()
	}

	if len(route.NodeID) == 0 {
		route.NodeID = l.allocateNode()
	}

	if len(route.NodeID) == 0 {
		return nil, fmt.Errorf("no enough node for user")
	}

	user = &model.User{
		UserName:             req.UserName,
		PasswordMD5:          passwordMD5,
		StartTime:            trafficLimit.StartTime,
		EndTime:              trafficLimit.EndTime,
		TotalTraffic:         trafficLimit.TotalTraffic * trafficUnit,
		RouteMode:            route.Mode,
		RouteNodeID:          route.NodeID,
		UpdateRouteIntervals: route.Intervals,
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

	err = model.BindNode(l.svcCtx.Redis, route.NodeID, req.UserName)
	if err != nil {
		return nil, err
	}

	node, err := model.GetNode(l.svcCtx.Redis, route.NodeID)
	if err != nil {
		return nil, err
	}

	createUserResp := &types.CreateUserResp{
		UserName:     req.UserName,
		TrafficLimit: trafficLimit,
		Route:        route,
		NodeIP:       node.IP,
	}

	return createUserResp, nil
}

func (l *CreateUserLogic) defaultRoute() *types.Route {
	return &types.Route{Mode: routeModeTypeManual, NodeID: l.allocateNode(), Intervals: 0}
}

func (l *CreateUserLogic) allocateNode() string {
	nodeID, err := model.GetOnlineAndUnbindNode(l.svcCtx.Redis)
	if err != nil {
		logx.Errorf("GetOnlineAndUnbindNode %v", err)
		return ""
	}
	return nodeID
}

func (l *CreateUserLogic) defaultTrafficLimit() *types.TrafficLimit {
	return &types.TrafficLimit{
		StartTime:    time.Now().Unix(),
		EndTime:      time.Now().AddDate(0, 1, 0).Unix(),
		TotalTraffic: defaultTotalTraffic,
	}
}
