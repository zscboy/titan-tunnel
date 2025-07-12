package logic

import (
	"context"
	"fmt"

	"titan-tunnel/server/api/model"
	"titan-tunnel/server/rpc/internal/svc"
	"titan-tunnel/server/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteUserLogic) DeleteUser(in *pb.DeleteUserReq) (*pb.UserOperationResp, error) {
	user, err := model.GetUser(l.svcCtx.Redis, in.UserName)
	if err != nil {
		return &pb.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if user == nil {
		return &pb.UserOperationResp{ErrMsg: fmt.Sprintf("user %s not exist", in.UserName)}, nil
	}

	err = model.DeleteUser(l.svcCtx.Redis, in.UserName)
	if err != nil {
		return &pb.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if len(user.RouteNodeID) > 0 {
		if err := model.UnbindNode(l.svcCtx.Redis, user.RouteNodeID); err != nil {
			return nil, err
		}
	}

	// l.svcCtx.TunMgr.DeleteUserFromCache(in.UserName)
	// TODO: notify api service update user cache
	return &pb.UserOperationResp{Success: true}, nil
}
