package logic

import (
	"context"
	"fmt"

	"titan-tunnel/server/api/model"
	"titan-tunnel/server/rpc/internal/svc"
	"titan-tunnel/server/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SwitchUserRouteNodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSwitchUserRouteNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SwitchUserRouteNodeLogic {
	return &SwitchUserRouteNodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SwitchUserRouteNodeLogic) SwitchUserRouteNode(in *pb.SwitchUserRouteNodeReq) (*pb.UserOperationResp, error) {
	if len(in.NodeId) != 0 {
		if err := l.checkNode(in.NodeId); err != nil {
			return &pb.UserOperationResp{ErrMsg: err.Error()}, nil
		}
	} else {
		nodeID, err := l.allocateNode()
		if err != nil {
			return &pb.UserOperationResp{ErrMsg: err.Error()}, nil
		}
		in.NodeId = nodeID
	}

	user, err := model.GetUser(l.svcCtx.Redis, in.UserName)
	if err != nil {
		return &pb.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if user == nil {
		return &pb.UserOperationResp{ErrMsg: fmt.Sprintf("user %s not exist", in.UserName)}, nil
	}

	oldNodeID := user.RouteNodeID
	user.RouteNodeID = in.NodeId

	err = model.SaveUser(l.svcCtx.Redis, user)
	if err != nil {
		return &pb.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	err = model.BindNode(l.svcCtx.Redis, user.RouteNodeID, user.UserName)
	if err != nil {
		return &pb.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if len(oldNodeID) > 0 {
		err = model.UnbindNode(l.svcCtx.Redis, oldNodeID)
		if err != nil {
			logx.Errorf("UnbindNode %s failed:%v", oldNodeID, err)
		}
	}

	// l.svcCtx.TunMgr.DeleteUserFromCache(in.UserName)
	// TODO: update user cache
	return &pb.UserOperationResp{Success: true}, nil
}

func (l *SwitchUserRouteNodeLogic) checkNode(nodeID string) error {
	node, err := model.GetNode(l.svcCtx.Redis, nodeID)
	if err != nil {
		return err
	}

	if node == nil {
		return fmt.Errorf("node %s not exist", nodeID)
	}

	if len(node.BindUser) != 0 {
		return fmt.Errorf("node %s alreay used by user %s", nodeID, node.BindUser)
	}

	if !node.Online {
		return fmt.Errorf("node %s offline", nodeID)
	}

	return nil
}
func (l *SwitchUserRouteNodeLogic) allocateNode() (string, error) {
	nodeID, err := model.GetOnlineAndUnbindNode(l.svcCtx.Redis)
	if err != nil {
		logx.Errorf("GetOnlineAndUnbindNode %v", err)
		return "", err
	}
	return nodeID, nil
}
