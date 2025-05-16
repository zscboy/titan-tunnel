package logic

import (
	"fmt"
	"strconv"
	"titan-vm/vms/virt"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func getVirtOpts(redis *redis.Redis, id string) (*virt.VirtOptions, error) {
	key := fmt.Sprintf("vms:host:%s", id)
	results, err := redis.Hmget(key, "os", "vmapi", "online")
	if err != nil {
		return nil, err
	}

	online, err := strconv.ParseBool(results[2])
	if err != nil {
		return nil, err
	}

	opts := &virt.VirtOptions{OS: results[0], VMAPI: results[1], Online: online}
	return opts, nil
}
