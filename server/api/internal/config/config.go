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
	UDPServerIP  string
	UDPPortStart int
	UDPPortEnd   int
	EnableAuth   bool
	TCPTimeout   int64
	UDPTimeout   int64
}

type Pop struct {
	ID         string
	Area       string
	Socks5Addr string
}

type Config struct {
	rest.RestConf
	Redis   redis.RedisConf
	JwtAuth JwtAuth
	Socks5  Socks5
	PopID   string
	// todo: will move to center server
	ServerURL string
	// todo: will move to center server
	Pops []Pop
}
