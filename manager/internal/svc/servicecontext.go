package svc

import (
	"context"
	"titan-tunnel/manager/internal/config"
	"titan-tunnel/server/rpc/serverapi"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Server struct {
	API         serverapi.ServerAPI
	Socks5Addr  string
	WSServerURL string
	Area        string
}

type ServiceContext struct {
	Config        config.Config
	Redis         *redis.Redis
	JwtMiddleware rest.Middleware
	Servers       map[string]*Server
}

func NewServiceContext(c config.Config) *ServiceContext {
	redis := redis.MustNewRedis(c.Redis)
	return &ServiceContext{
		Config:  c,
		Redis:   redis,
		Servers: newServers(c),
	}
}

// TODO: can not get server info in here, server may be stop
func newServers(c config.Config) map[string]*Server {
	servers := make(map[string]*Server)
	for _, pop := range c.Pops {
		api := serverapi.NewServerAPI(zrpc.MustNewClient(pop.RpcClient))
		resp, err := api.GetServerInfo(context.Background(), &serverapi.Empty{})
		if err != nil {
			panic("Get server info failed:" + err.Error())
		}
		servers[pop.Id] = &Server{API: api, Socks5Addr: resp.Socks5Addr, WSServerURL: resp.WsServerUrl, Area: pop.Area}
	}
	return servers
}
