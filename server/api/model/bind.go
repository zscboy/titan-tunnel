package model

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func BindNode(redis *redis.Redis, nodeID, userName string) error {
	node, err := GetNode(redis, nodeID)
	if err != nil {
		return err
	}

	if node == nil {
		return fmt.Errorf("node %s not exist", nodeID)
	}

	node.BindUser = userName
	if err = SaveNode(redis, node); err != nil {
		return err
	}

	t, err := parseTimeFromString(node.RegisterAt)
	if err != nil {
		return err
	}

	_, err = redis.Zadd(redisKeyNodeBind, t.Unix(), nodeID)
	if err != nil {
		return err
	}

	_, err = redis.Zrem(redisKeyNodeFree, nodeID)
	if err != nil {
		return err
	}

	return nil
}

func UnbindNode(redis *redis.Redis, nodeID string) error {
	node, err := GetNode(redis, nodeID)
	if err != nil {
		return err
	}

	if node == nil {
		return fmt.Errorf("node %s not exist", nodeID)
	}

	node.BindUser = ""
	if err = SaveNode(redis, node); err != nil {
		return err
	}

	_, err = redis.Zrem(redisKeyNodeBind, nodeID)
	if err != nil {
		return err
	}

	isOnline, err := isNodeOnline(redis, nodeID)
	if err != nil {
		return err
	}

	if isOnline {
		t, err := parseTimeFromString(node.LoginAt)
		if err != nil {
			return err
		}

		_, err = redis.Zadd(redisKeyNodeFree, t.Unix(), nodeID)
		if err != nil {
			return err
		}
	}

	return nil
}

func RemoveFreeNode(redis *redis.Redis, nodeID string) error {
	_, err := redis.Zrem(redisKeyNodeFree, nodeID)
	return err
}

func AddFreeNode(redis *redis.Redis, loginTimestamp int64, nodeID string) error {
	_, err := redis.Zadd(redisKeyNodeFree, loginTimestamp, nodeID)
	return err
}

func AllocateFreeNode(redis *redis.Redis) (string, error) {
	ids, err := redis.Zrange(redisKeyNodeFree, 0, 0)
	if err != nil {
		return "", err
	}

	if len(ids) > 0 {
		return ids[0], nil
	}
	return "", fmt.Errorf("no free node found")
}
