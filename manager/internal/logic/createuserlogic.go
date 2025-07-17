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
	server := l.svcCtx.Servers[req.PopId]
	if server == nil {
		return nil, fmt.Errorf("pop %s not found", req.PopId)
	}

	in := &serverapi.CreateUserReq{UserName: req.UserName, Password: req.Password}
	if req.TrafficLimit != nil {
		in.TrafficLimit = toTrafficLimitReq(req.TrafficLimit)
	}

	if req.Route != nil {
		in.Route = toRouteReq(req.Route)
	}

	createUserResp, err := server.API.CreateUser(l.ctx, in)
	if err != nil {
		return nil, err
	}

	if err := model.SaveUser(l.svcCtx.Redis, &model.User{UserName: req.UserName, PopID: req.PopId}); err != nil {
		return nil, err
	}

	return toCreateUserResp(createUserResp), nil
}
