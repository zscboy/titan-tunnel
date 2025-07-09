package logic

import (
	"fmt"
	"time"
	"titan-tunnel/server/internal/types"
	"titan-tunnel/server/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	routeModeType = iota
	routeModeTypeManual
	routeModeTypeAuto
	routeModeTypeTimed
)

func checkRoute(redis *redis.Redis, route *types.Route) error {
	if route == nil {
		return fmt.Errorf("route is empty")
	}
	if isInvalidRouteMode(route.Mode) {
		return fmt.Errorf("invalid route mode %d", route.Mode)
	}

	// check node if exist or used by other user
	if len(route.NodeID) != 0 {
		node, err := model.GetNode(redis, route.NodeID)
		if err != nil {
			return err
		}

		if node == nil {
			return fmt.Errorf("node %s not exist", route.NodeID)
		}

		if len(node.BindUser) != 0 {
			return fmt.Errorf("node %s alreay used by user %s", route.NodeID, node.BindUser)
		}
	}

	return nil
}

func isInvalidRouteMode(mode int) bool {
	if mode != routeModeTypeManual && mode != routeModeTypeAuto && mode != routeModeTypeTimed {
		return false
	}

	return true
}

func checkTraffic(trafficLimit *types.TrafficLimit) error {
	if trafficLimit.EndTime <= trafficLimit.StartTime {
		return fmt.Errorf("invalid traffic start time %d and end time %d", trafficLimit.StartTime, trafficLimit.EndTime)
	}

	if trafficLimit.EndTime < time.Now().Unix() {
		return fmt.Errorf("traffic end time is out of date", trafficLimit.EndTime)
	}

	if trafficLimit.TotalTraffic <= 0 {
		return fmt.Errorf("invalid total traffic ", trafficLimit.TotalTraffic)
	}

	return nil
}
