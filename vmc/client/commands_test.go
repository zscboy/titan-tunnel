package client

import (
	"testing"
	"titan-vm/pb"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func TestProto(t *testing.T) {
	request := &pb.CmdDownloadImageRequest{Url: "http://baidu.com"}
	bytes, err := proto.Marshal(request)
	if err != nil {
		t.Fatal(err.Error())
	}

	cmd := &pb.Command{Type: pb.CommandType_DownloadImage, Data: bytes}
	payloadData, err := proto.Marshal(cmd)
	if err != nil {
		t.Fatal(err.Error())
	}

	msg := &pb.Message{Type: pb.MessageType_CONTROL, SessionId: uuid.NewString(), Payload: payloadData}
	msgBuf, err := proto.Marshal(msg)
	if err != nil {
		t.Fatal(err.Error())
	}

	msg2 := &pb.Message{}
	err = proto.Unmarshal(msgBuf, msg2)
	if err != nil {
		t.Fatal(err.Error())
	}

	cmd2 := &pb.Command{}
	err = proto.Unmarshal(msg2.GetPayload(), cmd2)
	if err != nil {
		t.Fatal(err.Error())
	}

	request2 := &pb.CmdDownloadImageRequest{}
	err = proto.Unmarshal(cmd2.GetData(), request2)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("%#v", request2)

}
