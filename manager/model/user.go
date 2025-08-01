package model

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func SetUserPop(redis *redis.Redis, user, pop string) error {
	return redis.Hset(redisKeyUsers, user, pop)
}

func GetUserPop(red *redis.Redis, user string) (string, error) {
	popID, err := red.Hget(redisKeyUsers, user)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}

		return "", err
	}

	return popID, nil
}

func DeleteUser(redis *redis.Redis, user string) error {
	_, err := redis.Hdel(redisKeyUsers, user)
	return err
}
