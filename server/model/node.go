package model

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Node struct {
	Id      string
	OS      string `redis:"os"`
	VmAPI   string `redis:"vmapi"`
	CPU     int    `redis:"cpu"`
	Memory  int    `redis:"memory"`
	LoginAt string `redis:"loginAt"`
	// OfflineAt  string `redis:"offlineAt"`
	RegisterAt string `redis:"registerAt"`
	Online     bool
	IP         string `redis:"ip"`
	SSHPort    int    `redis:"sshPort"`
	PubKey     string `redis:"pubKey"`
	Extend     string `redis:"extend"`
}

func SetNodeWithZadd(ctx context.Context, redis *redis.Redis, node *Node) error {
	hashKey := fmt.Sprintf(redisKeyVmsNode, node.Id)
	m, err := structToMap(node)
	if err != nil {
		return err
	}

	layout := "2006-01-02 15:04:05 -0700 MST"
	t, err := time.Parse(layout, node.RegisterAt)
	if err != nil {
		return err
	}

	pipe, err := redis.TxPipeline()
	if err != nil {
		return err
	}

	pipe.HMSet(ctx, hashKey, m)
	pipe.ZAdd(ctx, redisKeyVmsZset, goredis.Z{Score: float64(t.Unix()), Member: node.Id})

	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return err
}

func SaveNode(redis *redis.Redis, node *Node) error {
	key := fmt.Sprintf(redisKeyVmsNode, node.Id)
	m, err := structToMap(node)
	if err != nil {
		return err
	}

	logx.Infof("m:%v", m)
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

	node := &Node{}
	err = mapToStruct(m, node)
	if err != nil {
		return nil, err
	}

	online, err := isNodeOnline(redis, id)
	if err != nil {
		return nil, err
	}

	node.Id = id
	node.Online = online
	return node, nil
}

func ListNode(ctx context.Context, redis *redis.Redis, start, end int) ([]*Node, error) {
	ids, err := redis.Zrevrange(redisKeyVmsZset, int64(start), int64(end))
	if err != nil {
		return nil, err
	}

	pipe1, err := redis.TxPipeline()
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		key := fmt.Sprintf(redisKeyVmsOnline, id)
		pipe1.Exists(ctx, key)
	}

	results, err := pipe1.Exec(ctx)
	if err != nil {
		return nil, err
	}

	onlines := make([]bool, 0, len(ids))
	for _, result := range results {
		exist, err := result.(*goredis.IntCmd).Result()
		if err != nil {
			logx.Errorf("ListNode parse result failed:%s", err.Error())
			continue
		}
		onlines = append(onlines, exist == 1)
	}

	pipe2, err := redis.TxPipeline()
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		key := fmt.Sprintf(redisKeyVmsNode, id)
		pipe2.HGetAll(ctx, key)
	}

	cmds, err := pipe2.Exec(ctx)
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

		node := Node{Id: ids[i], Online: onlines[i]}
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
	return redis.Zcard(redisKeyVmsZset)
}

func SetNodeOnline(redis *redis.Redis, nodeId string) error {
	key := fmt.Sprintf(redisKeyVmsOnline, nodeId)
	if err := redis.Set(key, "true"); err != nil {
		return err
	}

	return redis.Expire(key, 60)
}

func SetNodeOffline(redis *redis.Redis, nodeId string) error {
	key := fmt.Sprintf(redisKeyVmsOnline, nodeId)
	_, err := redis.Del(key)
	return err
}

func isNodeOnline(redis *redis.Redis, nodeId string) (bool, error) {
	key := fmt.Sprintf(redisKeyVmsOnline, nodeId)
	return redis.Exists(key)
}
