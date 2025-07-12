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
	api := l.svcCtx.ServerAPIs[req.PopId]
	if api == nil {
		return nil, fmt.Errorf("pop %s not found", req.PopId)
	}

	//   UserName     string        `json:"user_name"`
	// Password     string        `json:"password"`
	// PopId        string        `json:"pop_id"`
	// TrafficLimit *TrafficLimit `json:"traffic_limit,optional"`
	// Route        *Route        `json:"route,optional"`
	in := &serverapi.CreateUserReq{UserName: req.UserName, Password: req.Password, PopId: req.PopId}
	if req.TrafficLimit != nil {
		in.TrafficLimit = toTrafficLimitReq(req.TrafficLimit)
	}

	if req.Route != nil {
		in.Route = toRouteReq(req.Route)
	}

	createUserResp, err := api.CreateUser(l.ctx, in)
	if err != nil {
		return nil, err
	}

	if err := model.SaveUser(l.svcCtx.Redis, &model.User{UserName: req.UserName, PopID: req.PopId}); err != nil {
		return nil, err
	}

	return toCreateUserResp(createUserResp), nil
}
