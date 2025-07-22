package model

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type User struct {
	UserName string
	PopID    string `redis:"pop_id"`
}

func SaveUser(redis *redis.Redis, user *User) error {
	m, err := structToMap(user)
	if err != nil {
		return err
	}

	key := fmt.Sprintf(redisKeyUser, user.UserName)
	return redis.Hmset(key, m)
}

// if user not exist, will return nil, nil
func GetUser(redis *redis.Redis, userName string) (*User, error) {
	key := fmt.Sprintf(redisKeyUser, userName)
	m, err := redis.Hgetall(key)
	if err != nil {
		return nil, err
	}

	if len(m) == 0 {
		return nil, nil
	}

	user := &User{}
	err = mapToStruct(m, user)
	if err != nil {
		return nil, err
	}
	user.UserName = userName

	return user, nil
}

func DeleteUser(redis *redis.Redis, userName string) error {
	key := fmt.Sprintf(redisKeyUser, userName)
	_, err := redis.Del(key)
	return err
}
