package export

import (
	"titan-tunnel/server/rpc/internal/config"
	"titan-tunnel/server/rpc/internal/server"
	"titan-tunnel/server/rpc/internal/svc"
	"titan-tunnel/server/rpc/pb"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type RPCServerConfig config.Config

func AddRPCService(group *service.ServiceGroup, c RPCServerConfig) {
	ctx := svc.NewServiceContext(config.Config(c))

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterServerAPIServer(grpcServer, server.NewServerAPIServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	group.Add(s)
}
