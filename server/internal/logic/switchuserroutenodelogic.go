package logic

import (
	"context"
	"fmt"

	"titan-tunnel/server/internal/svc"
	"titan-tunnel/server/internal/types"
	"titan-tunnel/server/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type SwitchUserRouteNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSwitchUserRouteNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SwitchUserRouteNodeLogic {
	return &SwitchUserRouteNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SwitchUserRouteNodeLogic) SwitchUserRouteNode(req *types.SwitchUserRouteNodeReq) (resp *types.UserOperationResp, err error) {
	if len(req.NodeId) != 0 {
		if err = l.checkNode(req.NodeId); err != nil {
			return &types.UserOperationResp{ErrMsg: err.Error()}, nil
		}
	} else {
		nodeID, err := l.allocateNode()
		if err != nil {
			return &types.UserOperationResp{ErrMsg: err.Error()}, nil
		}
		req.NodeId = nodeID
	}

	user, err := model.GetUser(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if user == nil {
		return &types.UserOperationResp{ErrMsg: fmt.Sprintf("user %s not exist", req.UserName)}, nil
	}

	oldNodeID := user.RouteNodeID
	user.RouteNodeID = req.NodeId

	err = model.SaveUser(l.svcCtx.Redis, user)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	err = model.BindNode(l.svcCtx.Redis, user.RouteNodeID)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if len(oldNodeID) > 0 {
		err = model.UnbindNode(l.svcCtx.Redis, oldNodeID)
		if err != nil {
			logx.Errorf("UnbindNode %s failed:%v", oldNodeID, err)
		}
	}

	l.svcCtx.TunMgr.DeleteUserFromCache(req.UserName)
	return &types.UserOperationResp{Success: true}, nil
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
