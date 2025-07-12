package model

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func BindNode(redis *redis.Redis, nodeID, userName string) error {
	_, err := redis.Zrem(redisKeyNodeUnbind, nodeID)
	if err != nil {
		return err
	}
	_, err = redis.Zadd(redisKeyNodeBind, time.Now().Unix(), nodeID)
	if err != nil {
		return err
	}

	node, err := GetNode(redis, nodeID)
	if err != nil {
		return err
	}

	if node.BindUser != userName {
		node.BindUser = userName
		return SaveNode(redis, node)
	}

	return nil
}

func UnbindNode(redis *redis.Redis, nodeID string) error {
	_, err := redis.Zrem(redisKeyNodeBind, nodeID)
	if err != nil {
		return err
	}
	_, err = redis.Zadd(redisKeyNodeUnbind, time.Now().Unix(), nodeID)
	if err != nil {
		return err
	}

	node, err := GetNode(redis, nodeID)
	if err != nil {
		return err
	}

	if len(node.BindUser) > 0 {
		node.BindUser = ""
		return SaveNode(redis, node)
	}

	return nil
}

func getUnbindNodes(redis *redis.Redis, start, stop int64) ([]string, error) {
	ids, err := redis.Zrange(redisKeyNodeUnbind, start, stop)
	if err != nil {
		return nil, err
	}

	if len(ids) > 0 {
		return ids, nil
	}
	return nil, nil
}

func GetOnlineAndUnbindNode(redis *redis.Redis) (string, error) {
	start := int64(0)
	count := int64(20)
	for {
		nodes, err := getUnbindNodes(redis, start, start+count-1)
		if err != nil {
			return "", err
		}

		if len(nodes) == 0 {
			return "", nil
		}

		for _, nodeID := range nodes {
			online, err := isNodeOnline(redis, nodeID)
			if err != nil {
				return "", err
			}

			if online {
				return nodeID, nil
			}
		}

		start = start + count
	}
}
