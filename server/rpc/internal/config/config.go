package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	APIServer string   `json:",optional"`
	Whitelist []string `json:",optional"`
}
