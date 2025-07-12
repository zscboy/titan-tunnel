package svc

import (
	"titan-tunnel/manager/internal/config"
	"titan-tunnel/server/rpc/serverapi"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config        config.Config
	Redis         *redis.Redis
	JwtMiddleware rest.Middleware
	ServerAPIs    map[string]serverapi.ServerAPI
}

func NewServiceContext(c config.Config) *ServiceContext {
	redis := redis.MustNewRedis(c.Redis)
	return &ServiceContext{
		Config:     c,
		Redis:      redis,
		ServerAPIs: newServerAPIS(c),
	}
}

func newServerAPIS(c config.Config) map[string]serverapi.ServerAPI {
	apis := make(map[string]serverapi.ServerAPI)
	for _, pop := range c.Pops {
		api := serverapi.NewServerAPI(zrpc.MustNewClient(pop.RpcClient))
		apis[pop.ID] = api
	}
	return apis
}
