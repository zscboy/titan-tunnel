package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type JwtAuth struct {
	AccessSecret string
	AccessExpire int64
}

type Config struct {
	zrpc.RpcServerConf
	RestConf      rest.RestConf
	RpcClient     zrpc.RpcClientConf
	MultipassCert string
	MultipassKey  string
	JwtAuth       JwtAuth
	SSHPriKey     string
	SSHPubKey     string
}
