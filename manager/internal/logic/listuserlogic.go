package logic

import (
	"context"
	"fmt"

	"titan-tunnel/manager/internal/svc"
	"titan-tunnel/manager/internal/types"
	"titan-tunnel/server/rpc/serverapi"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserLogic {
	return &ListUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserLogic) ListUser(req *types.ListUserReq) (resp *types.ListUserResp, err error) {
	server := l.svcCtx.Servers[req.PopID]
	if server == nil {
		return nil, fmt.Errorf("pop %s not found", req.PopID)
	}

	in := &serverapi.ListUserReq{
		Start: int32(req.Start),
		End:   int32(req.End),
	}

	listUserResp, err := server.API.ListUser(l.ctx, in)
	if err != nil {
		return nil, err
	}

	users := make([]*types.User, 0, len(listUserResp.Users))
	for _, user := range listUserResp.Users {
		u := &types.User{
			UserName:       user.UserName,
			NodeIP:         user.NodeIp,
			NodeOnline:     user.NodeOnline,
			CurrentTraffic: user.CurrentTraffic,
			Off:            user.Off,
			TrafficLimit:   toTrafficLimitResp(user.TrafficLimit),
			Route:          toRouteResp(user.Route),
		}
		users = append(users, u)
	}

	return &types.ListUserResp{Users: users, Total: int(listUserResp.Total)}, nil
}
