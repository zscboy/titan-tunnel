package model

const (
	redisKeyUser     = "titan:user:%s"
	redisKeyUserZset = "titan:user:zset"

	redisKeyNode     = "titan:node:%s"
	redisKeyNodeZset = "titan:node:zset"
	// key expire
	redisKeyNodeOnline = "titan:node:online:%s"
	// sort set
	redisKeyNodeBind = "titan:node:bind"
	// sort set
	redisKeyNodeUnbind = "titan:node:unbind"

	redisKeyBrowser       = "titan:browser:%s"
	redisKeyBrowserZset   = "titan:browser:zset"
	redisKeyBrowserOnline = "titan:browser:online:%s"
	redisKeyBrowserBind   = "titan:browser:bind"
	redisKeyBrowserUnbind = "titan:browser:unbind"
)
