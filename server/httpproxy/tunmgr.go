package httpproxy

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"titan-tunnel/server/api/model"

	"github.com/bluele/gcache"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	userCacheSize = 512
)

type TunnelManager struct {
	tunnels     sync.Map
	redis       *redis.Redis
	userTraffic *userTraffic
	userCache   gcache.Cache
}

func NewTunnelManager(redis *redis.Redis) *TunnelManager {
	tm := &TunnelManager{
		redis:       redis,
		userTraffic: newUserTraffic(),
		userCache:   gcache.New(userCacheSize).LRU().Build(),
	}
	go tm.keepalive()
	go tm.startUserTrafficTimer()
	return tm
}

func (tm *TunnelManager) acceptWebsocket(conn *websocket.Conn, browser *model.Browser) {
	logx.Debugf("TunnelManager:%s accept websocket ", browser.Id)
	v, ok := tm.tunnels.Load(browser.Id)
	if ok {
		oldTun := v.(*Tunnel)
		oldTun.close()
	}

	tun := newTunnel(conn, tm, &TunOptions{Id: browser.Id, OS: browser.OS, IP: browser.IP})

	tm.tunnels.Store(browser.Id, tun)
	defer tm.tunnels.Delete(browser.Id)

	if err := model.SetBrowserAndZadd(context.Background(), tm.redis, browser); err != nil {
		logx.Errorf("SetNode failed:%s", err.Error())
		return
	}

	if err := model.SetBrowserOnline(tm.redis, browser.Id); err != nil {
		logx.Errorf("SetNodeOnline failed:%s", err.Error())
		return
	}

	// if len(browser.BindUser) > 0 {
	// 	model.BindBrowser(tm.redis, browser.Id, browser.BindUser)
	// } else {
	// 	model.UnbindNode(tm.redis, browser.Id)
	// }

	defer model.SetBrowserOffline(tm.redis, browser.Id)

	tun.serve()
}

func (tm *TunnelManager) onHTTPRequest(targetInfo *TargetInfo) error {
	logx.Debugf("req: %v", *targetInfo.req)
	tun, err := tm.randomTrunnelForUser(targetInfo.userName)
	if err != nil {
		return err
	}

	return tun.onHTTPRequest(targetInfo)
}

func (tm *TunnelManager) randomTrunnelForUser(userName string) (*Tunnel, error) {
	var t *Tunnel
	tm.tunnels.Range(func(key, value any) bool {
		t = value.(*Tunnel)
		return false
	})

	if t == nil {
		return nil, fmt.Errorf("can not find tunnel")
	}

	return t, nil
}
func (tm *TunnelManager) getUserFromCache(userName string) (*model.User, error) {
	v, err := tm.userCache.Get(userName)
	if err != nil {
		logx.Infof("getUserFromCache:%v", err)
		if !errors.Is(err, gcache.KeyNotFoundError) {
			return nil, err
		}

		user, err := model.GetUser(tm.redis, userName)
		if err != nil {
			return nil, err
		}

		if user == nil {
			return nil, fmt.Errorf("user %s not exist", userName)
		}

		tm.userCache.Set(userName, user)

		return user, nil
	}
	logx.Infof("getUserFromCache:%v", v)
	return v.(*model.User), nil
}

func (tm *TunnelManager) getTunnelByUser(userName string) (*Tunnel, error) {
	user, err := tm.getUserFromCache(userName)
	if err != nil {
		return nil, err
	}

	nodeID := user.RouteNodeID
	return tm.getTunnel(nodeID), nil

}

func (tm *TunnelManager) DeleteUserFromCache(userName string) {
	tm.userCache.Remove(userName)
}

func (tm *TunnelManager) keepalive() {
	tick := 0
	for {
		time.Sleep(time.Second * 1)
		tick++

		if tick == 30 {
			tick = 0
			tm.tunnels.Range(func(key, value any) bool {
				t := value.(*Tunnel)
				t.keepalive()
				return true
			})
		}
	}
}

func (tm *TunnelManager) startUserTrafficTimer() {
	ticker := time.NewTicker(300 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		trafficMap := tm.userTraffic.snapshotAndClear()
		for userName, traffic := range trafficMap {
			if traffic > 0 {
				user, err := model.GetUser(tm.redis, userName)
				if err != nil {
					logx.Errorf("get user %v", err)
					continue
				}

				if user == nil {
					logx.Errorf("user %s not exist", userName)
					continue
				}

				user.CurrentTraffic += traffic
				model.SaveUser(tm.redis, user)
			}
		}
	}
}

func (tm *TunnelManager) getTunnel(id string) *Tunnel {
	v, ok := tm.tunnels.Load(id)
	if !ok {
		return nil
	}
	return v.(*Tunnel)
}

func (tm *TunnelManager) traffic(userName string, traffic int64) {
	tm.userTraffic.add(userName, traffic)
}

func (tm *TunnelManager) getTrafficAndDelete(userName string) int64 {
	return tm.userTraffic.getAndDelete(userName)
}
