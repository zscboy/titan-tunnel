package model

import (
	goredis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Account struct {
	UserName    string `redis:"userName"`
	PasswordMD5 string `redis:"password_md5"`
}

func SaveAccount(redis *redis.Redis, account *Account) error {
	m, err := structToMap(account)
	if err != nil {
		return err
	}
	return redis.Hmset(redisKeyUserAccount, m)
}

func GetAccount(redis *redis.Redis, userName string) (*Account, error) {
	passordMD5, err := redis.Hget(redisKeyUserAccount, userName)
	if err != nil {
		if err == goredis.Nil {
			return nil, nil
		}
		return nil, err
	}

	return &Account{UserName: userName, PasswordMD5: passordMD5}, nil
}
