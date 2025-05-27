package logic

import (
	"fmt"
	"titan-vm/vms/model"
	"titan-vm/vms/virt"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func getVirtOpts(redis *redis.Redis, id string) (*virt.VirtOptions, error) {
	node, err := model.GetNode(redis, id)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, fmt.Errorf("not found %s", id)
	}

	opts := &virt.VirtOptions{OS: node.OS, VMAPI: node.VmAPI, Online: node.Online}
	return opts, nil
}
