package proto

import (
	"testing"

	"github.com/google/uuid"
)

func TestProto(t *testing.T) {
	addr := &DestAddr{Network: "tcp", Addr: "localhost:8080"}
	buf, err := addr.EncodeMessage()
	if err != nil {
		t.Errorf("encode addr %s", err.Error())
		return
	}

	msg := &Message{
		Type:      MsgTypeProxySessionCreate,
		SessionID: uuid.NewString(),
		Payload:   buf,
	}

	buf, err = msg.EncodeMessage()
	if err != nil {
		t.Errorf("encode message %s", err.Error())
		return
	}

	decodeMsg := &Message{}
	err = decodeMsg.DecodeMessage(buf)
	if err != nil {
		t.Errorf("decode message %s", err.Error())
		return
	}

	t.Logf("msg type:%d, session %s", decodeMsg.Type, decodeMsg.SessionID)

	decodeAddr := DestAddr{}
	decodeAddr.DecodeMessage(decodeMsg.Payload)
	if err != nil {
		t.Errorf("decode addr %s", err.Error())
		return
	}

	t.Logf("addr network:%s, addr %s", decodeAddr.Network, decodeAddr.Addr)

}
