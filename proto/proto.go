package proto

import (
	"bytes"
	"encoding/binary"
)

// encode 编码消息为binary
// 定义消息类型
type MessageType uint16

const (
	MsgTypeUnknown MessageType = iota
	// 命令类型的消息
	MsgTypeControl
	// 创建代理连接
	MsgTypeProxySessionCreate
	// 代理转发数据
	MsgTypeProxySessionData
	// 代理连接关闭
	MsgTypeProxySessionClose
)

func (mt MessageType) String() string {
	switch mt {
	case MsgTypeControl:
		return "Control"
	case MsgTypeProxySessionCreate:
		return "ProxySessionCreate"
	case MsgTypeProxySessionData:
		return "MsgTypeProxySessionData"
	case MsgTypeProxySessionClose:
		return "MsgTypeProxySessionClose"
	default:
		return "Unknown"
	}
}

// 定义消息体
type Message struct {
	Type      MessageType
	SessionID string
	Payload   []byte
}

// 命令类型消息的消息体
type CtrlPayload struct {
	Command string
	Args    []string
}

func (msg *Message) EncodeMessage() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.LittleEndian, msg.Type); err != nil {
		return nil, err
	}

	sessionIDBytes := []byte(msg.SessionID)
	if err := binary.Write(buf, binary.LittleEndian, uint16(len(sessionIDBytes))); err != nil {
		return nil, err
	}
	if _, err := buf.Write(sessionIDBytes); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.LittleEndian, uint32(len(msg.Payload))); err != nil {
		return nil, err
	}

	if len(msg.Payload) > 0 {
		if _, err := buf.Write(msg.Payload); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (msg *Message) DecodeMessage(data []byte) error {
	buf := bytes.NewReader(data)

	var msgType uint16
	if err := binary.Read(buf, binary.LittleEndian, &msgType); err != nil {
		return err
	}
	msg.Type = MessageType(msgType)

	var sessionIDLen uint16
	if err := binary.Read(buf, binary.LittleEndian, &sessionIDLen); err != nil {
		return err
	}

	sessionID := make([]byte, sessionIDLen)
	if _, err := buf.Read(sessionID); err != nil {
		return err
	}
	msg.SessionID = string(sessionID)

	var payloadLen uint32
	if err := binary.Read(buf, binary.LittleEndian, &payloadLen); err != nil {
		return err
	}

	if payloadLen > 0 {
		payload := make([]byte, payloadLen)
		if _, err := buf.Read(payload); err != nil {
			return err
		}
		msg.Payload = payload
	}

	return nil
}

type DestAddr struct {
	Network string
	Addr    string
}

func (destAddr *DestAddr) EncodeMessage() ([]byte, error) {
	buf := new(bytes.Buffer)

	addrBytes := []byte(destAddr.Addr)
	if err := binary.Write(buf, binary.LittleEndian, uint16(len(addrBytes))); err != nil {
		return nil, err
	}
	if _, err := buf.Write(addrBytes); err != nil {
		return nil, err
	}

	networkBytes := []byte(destAddr.Network)
	if err := binary.Write(buf, binary.LittleEndian, uint16(len(networkBytes))); err != nil {
		return nil, err
	}
	if _, err := buf.Write(networkBytes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (destAddr *DestAddr) DecodeMessage(data []byte) error {
	buf := bytes.NewReader(data)

	var addrLen uint16
	if err := binary.Read(buf, binary.LittleEndian, &addrLen); err != nil {
		return err
	}

	addr := make([]byte, addrLen)
	if _, err := buf.Read(addr); err != nil {
		return err
	}
	destAddr.Addr = string(addr)

	var networkLen uint16
	if err := binary.Read(buf, binary.LittleEndian, &networkLen); err != nil {
		return err
	}

	network := make([]byte, networkLen)
	if _, err := buf.Read(network); err != nil {
		return err
	}
	destAddr.Network = string(network)

	return nil
}
