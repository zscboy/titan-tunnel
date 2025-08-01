package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type JwtAuth struct {
	AccessSecret string
	AccessExpire int64
}

type Socks5 struct {
	Addr         string
	ServerIP     string
	UDPPortStart int
	UDPPortEnd   int
	EnableAuth   bool
	TCPTimeout   int64
	UDPTimeout   int64
}

type Config struct {
	rest.RestConf
	Redis   redis.RedisConf `json:",optional,inherit"`
	JwtAuth JwtAuth
	Socks5  Socks5
	Domain  string `json:",optional"`
}
