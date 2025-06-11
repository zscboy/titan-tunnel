package cmds

import (
	"testing"
)

func TestProto(t *testing.T) {
	hostInfo := NewHostInfo()
	resp := hostInfo.Get()
	t.Logf("%v", *resp)

}
