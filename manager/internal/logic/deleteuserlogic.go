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

type DeleteUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteUserLogic) DeleteUser(req *types.DeleteUserReq) (resp *types.UserOperationResp, err error) {
	user, err := model.GetUser(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	api := l.svcCtx.ServerAPIs[user.PopID]
	if api == nil {
		return &types.UserOperationResp{ErrMsg: fmt.Sprintf("pop %s not found", user.PopID)}, nil
	}

	deleteUserResp, err := api.DeleteUser(l.ctx, &serverapi.DeleteUserReq{UserName: user.UserName})
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	return &types.UserOperationResp{Success: deleteUserResp.Success, ErrMsg: deleteUserResp.ErrMsg}, nil
}
