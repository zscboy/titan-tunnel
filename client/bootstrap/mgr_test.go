package bootstrap

import (
	"testing"
)

func TestBootstrap(t *testing.T) {
	mgr, err := NewBootstrapMgr("./test")
	if err != nil {
		t.Log(err.Error())
		return
	}

	bootstraps := mgr.Bootstraps()
	t.Logf("bootstraps:%v", bootstraps)

}
