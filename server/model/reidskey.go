package model

const (
	redisKeyNode     = "titan:node:%s"
	redisKeyNodeZset = "titan:node:zset"
	redisKeyUser     = "titan:user:%s"
	redisKeyUserZset = "titan:user:zset"
	// key expire
	redisKeyNodeOnline = "titan:node:online:%s"
	// sort set
	redisKeyNodeBind = "titan:node:bind"
	// sort set
	redisKeyNodeUnbind = "titan:node:unbind"
)
