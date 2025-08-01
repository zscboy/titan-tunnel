package model

import (
	"testing"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func TestSaveUser(t *testing.T) {
	conf := redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}
	rd := redis.MustNewRedis(conf)
	// user := User{
	// 	UserName: "abc",
	// 	PopID:    "abc",
	// }

	userName := "abc"
	popID := "abc"
	if err := SetUserPop(rd, userName, popID); err != nil {
		t.Errorf("save user %v", err)
		return
	}
	t.Logf("save user success")
}

func TestGetUser(t *testing.T) {
	conf := redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}
	rd := redis.MustNewRedis(conf)
	user := "abc"
	pop, err := GetUserPop(rd, user)
	if err != nil {
		t.Errorf("get user %v", err)
		return
	}

	t.Logf("user:%s pop:%s", user, pop)
}
