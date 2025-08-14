package model

import (
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func SetNodePop(redis *redis.Redis, nodeID, pop string) error {
	if err := redis.Hset(redisKeyNodes, nodeID, pop); err != nil {
		return err
	}

	_, err := redis.Sadd(fmt.Sprintf(redisKeyPopNodes, pop), nodeID)
	return err
}

func GetNodePop(red *redis.Redis, nodeID string) ([]byte, error) {
	popID, err := red.Hget(redisKeyNodes, nodeID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, err
	}

	return []byte(popID), nil
}

func DeleteNode(redis *redis.Redis, nodeID string) error {
	popID, err := GetNodePop(redis, nodeID)
	if err != nil {
		return err
	}

	_, err = redis.Hdel(redisKeyNodes, nodeID)
	if err != nil {
		return err
	}

	_, err = redis.Srem(fmt.Sprintf(redisKeyPopNodes, popID), nodeID)
	return err
}

func NodeCountOfPop(redis *redis.Redis, popID string) (int64, error) {
	return redis.Scard(fmt.Sprintf(redisKeyPopNodes, popID))
}
