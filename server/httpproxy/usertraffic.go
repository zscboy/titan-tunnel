package httpproxy

import "sync"

type userTraffic struct {
	users map[string]int64
	lock  sync.Mutex
}

func newUserTraffic() *userTraffic {
	return &userTraffic{users: make(map[string]int64)}
}

func (ut *userTraffic) add(userName string, traffic int64) int64 {
	ut.lock.Lock()
	defer ut.lock.Unlock()

	ut.users[userName] = ut.users[userName] + traffic
	return ut.users[userName]
}

func (ut *userTraffic) getAndDelete(userName string) int64 {
	ut.lock.Lock()
	defer ut.lock.Unlock()

	traffic := ut.users[userName]
	defer delete(ut.users, userName)
	return traffic
}

func (ut *userTraffic) snapshotAndClear() map[string]int64 {
	ut.lock.Lock()
	defer ut.lock.Unlock()

	result := make(map[string]int64, len(ut.users))
	for user, traffic := range ut.users {
		result[user] = traffic
	}
	ut.users = make(map[string]int64) // 清空
	return result
}
