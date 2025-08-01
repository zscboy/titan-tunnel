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
	popID, err := model.GetUserPop(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return nil, err
	}

	if len(popID) == 0 {
		return nil, fmt.Errorf("user %s not exist", req.UserName)
	}

	server := l.svcCtx.Servers[popID]
	if server == nil {
		return nil, fmt.Errorf("pop %s not found", popID)
	}

	getUserResp, err := server.API.GetUser(l.ctx, &serverapi.GetUserReq{UserName: req.UserName})
	if err != nil {
		return nil, err
	}

	traffic := toTrafficLimitResp(getUserResp.TrafficLimit)
	route := toRouteResp(getUserResp.Route)

	return &types.GetUserResp{
		UserName:       getUserResp.UserName,
		PopId:          popID,
		NodeIP:         getUserResp.NodeIp,
		NodeOnline:     getUserResp.NodeOnline,
		CurrentTraffic: getUserResp.CurrentTraffic,
		Off:            getUserResp.Off,
		TrafficLimit:   traffic,
		Route:          route,
	}, nil
}
