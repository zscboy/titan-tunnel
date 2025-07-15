package config

import (
	api "titan-tunnel/server/api/export"
	rpc "titan-tunnel/server/rpc/export"
)

type Config struct {
	api.APIServerConfig
	rpc.RPCServerConfig
}
