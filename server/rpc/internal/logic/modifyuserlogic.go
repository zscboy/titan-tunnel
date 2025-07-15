package logic

import (
	"context"
	"fmt"

	"titan-tunnel/server/api/model"
	"titan-tunnel/server/rpc/internal/svc"
	"titan-tunnel/server/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewModifyUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyUserLogic {
	return &ModifyUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ModifyUserLogic) ModifyUser(in *pb.ModifyUserReq) (*pb.UserOperationResp, error) {
	if in.TrafficLimit == nil {
		return nil, fmt.Errorf("traffic limit not allow")
	}

	if in.Route == nil {
		return nil, fmt.Errorf("route not allow")
	}

	if err := checkRoute(l.svcCtx.Redis, in.Route); err != nil {
		return nil, err
	}

	if err := checkTraffic(in.TrafficLimit); err != nil {
		return nil, err
	}

	user, err := model.GetUser(l.svcCtx.Redis, in.UserName)
	if err != nil {
		return &pb.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if user == nil {
		return &pb.UserOperationResp{ErrMsg: fmt.Sprintf("user %s not exist", in.UserName)}, nil
	}

	user.StartTime = in.TrafficLimit.StartTime
	user.EndTime = in.TrafficLimit.EndTime
	user.TotalTraffic = in.TrafficLimit.TotalTraffic

	oldRouteNodeID := user.RouteNodeID

	user.RouteMode = int(in.Route.Mode)
	user.RouteNodeID = in.Route.NodeId
	user.UpdateRouteIntervals = int(in.Route.Intervals)

	err = model.SaveUser(l.svcCtx.Redis, user)
	if err != nil {
		return &pb.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if oldRouteNodeID != user.RouteNodeID {
		model.UnbindNode(l.svcCtx.Redis, oldRouteNodeID)
		model.BindNode(l.svcCtx.Redis, user.RouteNodeID, user.UserName)
	}

	deleteUserCacheLogic := NewDeleteUserCache(l.ctx, l.svcCtx)
	if err := deleteUserCacheLogic.DeleteUserCache(in.UserName); err != nil {

		return &pb.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	return &pb.UserOperationResp{Success: true}, nil
}
