package model

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Node struct {
	Id         string
	OS         string `redis:"os"`
	VmAPI      string `redis:"vmapi"`
	CPU        int    `redis:"cpu"`
	Memory     int    `redis:"memory"`
	LoginAt    string `redis:"loginAt"`
	OfflineAt  string `redis:"offlineAt"`
	RegisterAt string `redis:"registerAt"`
	Online     bool   `redis:"online"`
	IP         string `redis:"ip"`
	SSHPort    int    `redis:"sshPort"`
}

func RegisterNode(ctx context.Context, redis *redis.Redis, node *Node) error {
	hashKey := fmt.Sprintf(redisKeyVmsNode, node.Id)
	m, err := structToMap(node)
	if err != nil {
		return err
	}

	pipe, err := redis.TxPipeline()
	if err != nil {
		return err
	}

	pipe.HMSet(ctx, hashKey, m)
	pipe.LPush(ctx, redisKeyVmsList, node.Id)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return err
}

func SaveNode(redis *redis.Redis, host *Node) error {
	key := fmt.Sprintf(redisKeyVmsNode, host.Id)
	m, err := structToMap(host)
	if err != nil {
		return err
	}
	return redis.Hmset(key, m)
}

// GetNode if node not exist, return nil
func GetNode(redis *redis.Redis, id string) (*Node, error) {
	key := fmt.Sprintf(redisKeyVmsNode, id)
	m, err := redis.Hgetall(key)
	if err != nil {
		return nil, err
	}

	if len(m) == 0 {
		return nil, nil
	}

	host := &Node{}
	err = mapToStruct(m, host)
	if err != nil {
		return nil, err
	}

	host.Id = id
	return host, nil
}

func ListNode(ctx context.Context, redis *redis.Redis, start, end int) ([]*Node, error) {
	ids, err := redis.Lrange(redisKeyVmsList, start, end)
	if err != nil {
		return nil, err
	}

	pipe, err := redis.TxPipeline()
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		key := fmt.Sprintf(redisKeyVmsNode, id)
		pipe.HGetAll(ctx, key)
	}

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	nodes := make([]*Node, 0, len(cmds))
	for i, cmd := range cmds {
		result, err := cmd.(*goredis.MapStringStringCmd).Result()
		if err != nil {
			logx.Errorf("ListNode parse result failed:%s", err.Error())
			continue
		}

		node := Node{Id: ids[i]}
		err = mapToStruct(result, &node)
		if err != nil {
			logx.Errorf("ListNode mapToStruct error:%s", err.Error())
			continue
		}

		nodes = append(nodes, &node)
	}

	return nodes, nil
}

func GetNodeLen(redis *redis.Redis) (int, error) {
	return redis.Llen(redisKeyVmsList)
}
