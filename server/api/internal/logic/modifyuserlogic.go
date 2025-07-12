package logic

import (
	"context"
	"fmt"

	"titan-tunnel/server/api/internal/svc"
	"titan-tunnel/server/api/internal/types"
	"titan-tunnel/server/api/model"

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
	if req.TrafficLimit == nil {
		return nil, fmt.Errorf("traffic limit not allow")
	}

	if req.Route == nil {
		return nil, fmt.Errorf("route not allow")
	}

	if err := checkRoute(l.svcCtx.Redis, req.Route); err != nil {
		return nil, err
	}

	if err := checkTraffic(req.TrafficLimit); err != nil {
		return nil, err
	}

	user, err := model.GetUser(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if user == nil {
		return &types.UserOperationResp{ErrMsg: fmt.Sprintf("user %s not exist", req.UserName)}, nil
	}

	user.StartTime = req.TrafficLimit.StartTime
	user.EndTime = req.TrafficLimit.EndTime
	user.TotalTraffic = req.TrafficLimit.TotalTraffic

	oldRouteNodeID := user.RouteNodeID

	user.RouteMode = req.Route.Mode
	user.RouteNodeID = req.Route.NodeID
	user.UpdateRouteIntervals = req.Route.Intervals

	err = model.SaveUser(l.svcCtx.Redis, user)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if oldRouteNodeID != user.RouteNodeID {
		model.UnbindNode(l.svcCtx.Redis, oldRouteNodeID)
		model.BindNode(l.svcCtx.Redis, user.RouteNodeID, user.UserName)
	}

	l.svcCtx.TunMgr.DeleteUserFromCache(req.UserName)
	return &types.UserOperationResp{Success: true}, nil
}
