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

type ModifyUserPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewModifyUserPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyUserPasswordLogic {
	return &ModifyUserPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyUserPasswordLogic) ModifyUserPassword(req *types.ModifyUserPasswordReq) (resp *types.UserOperationResp, err error) {
	user, err := model.GetUser(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if user == nil {
		return &types.UserOperationResp{ErrMsg: fmt.Sprintf("user %s not exist", req.UserName)}, nil
	}

	server := l.svcCtx.Servers[user.PopID]
	if server == nil {
		return &types.UserOperationResp{ErrMsg: fmt.Sprintf("pop %s not found", user.PopID)}, nil
	}

	in := &serverapi.ModifyUserPasswordReq{UserName: user.UserName, NewPassword: req.NewPassword}
	modifyUserPasswordResp, err := server.API.ModifyUserPassword(l.ctx, in)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	return &types.UserOperationResp{Success: modifyUserPasswordResp.Success, ErrMsg: modifyUserPasswordResp.ErrMsg}, nil
}
