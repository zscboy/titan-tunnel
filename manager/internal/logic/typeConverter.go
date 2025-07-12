package logic

import (
	"titan-tunnel/manager/internal/types"
	"titan-tunnel/server/rpc/serverapi"
)

func toCreateUserResp(in *serverapi.CreateUserResp) *types.CreateUserResp {
	resp := &types.CreateUserResp{
		UserName: in.UserName,
		PopId:    in.PopId,
		NodeIP:   in.NodeIp,
	}

	if in.TrafficLimit != nil {
		resp.TrafficLimit = toTrafficLimitResp(in.TrafficLimit)
	}

	if in.Route != nil {
		resp.Route = toRouteResp(in.Route)
	}

	return resp
}

func toTrafficLimitResp(in *serverapi.TrafficLimit) *types.TrafficLimit {
	if in == nil {
		return nil
	}

	trafficLimit := types.TrafficLimit{
		StartTime:    in.StartTime,
		EndTime:      in.EndTime,
		TotalTraffic: in.TotalTraffic,
	}
	return &trafficLimit
}

func toTrafficLimitReq(in *types.TrafficLimit) *serverapi.TrafficLimit {
	if in == nil {
		return nil
	}

	trafficLimit := serverapi.TrafficLimit{
		StartTime:    in.StartTime,
		EndTime:      in.EndTime,
		TotalTraffic: in.TotalTraffic,
	}
	return &trafficLimit
}

func toRouteResp(in *serverapi.Route) *types.Route {
	if in == nil {
		return nil
	}

	route := types.Route{
		Mode:      int(in.Mode),
		NodeID:    in.NodeId,
		Intervals: int(in.Intervals),
	}
	return &route
}

func toRouteReq(in *types.Route) *serverapi.Route {
	if in == nil {
		return nil
	}

	route := serverapi.Route{
		Mode:      int32(in.Mode),
		NodeId:    in.NodeID,
		Intervals: int32(in.Intervals),
	}
	return &route
}

func toNodeResp(in *serverapi.Node) *types.Node {
	if in == nil {
		return nil
	}

	node := &types.Node{
		Id:       in.Id,
		IP:       in.Ip,
		NetDelay: int(in.NetDelay),
		BindUser: in.BindUser,
		Online:   in.Online,
	}
	return node
}
