package protocol

import (
	"encoding/binary"
	"errors"
	"github.com/raochq/ant/protocol/pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var InvalidHead = errors.New("invalid message head")
var InvalidID = errors.New("invalid message ID")
var InvalidMsg = errors.New("message not bind MsgID")

const (
	headSize = 8
)

type Head struct {
	//协议号
	MsgID uint32
	//消息流水号
	PacketNo uint16
	//回应的流水号
	ReplyNo uint16
	//协议体
	Body proto.Message
}

// 根据协议名称获取对应的协议id
func GetMsgIDByName(name string) uint32 {
	val := pb.MsgID_value["_"+name]
	return uint32(val)
}

// 根据协议id获取对应协议名称
func GetMsgNameByID(id int32) string {
	val := pb.MsgID_name[id]
	if val != "" {
		return val[1:]
	}
	return ""
}

// 根据协议id创建对应的消息
func NewProtoMessage(id uint32) proto.Message {
	if id <= 0 {
		return nil
	}
	msgName := GetMsgNameByID(int32(id))
	if msgName == "" {
		return nil
	}
	tp, err := protoregistry.GlobalTypes.FindMessageByURL(msgName)
	if err != nil {
		return nil
	}
	return tp.New().Interface()
}

func Unmarshal(data []byte) (*Head, error) {
	if len(data) < headSize {
		return nil, InvalidHead
	}
	head := &Head{}
	head.MsgID = binary.BigEndian.Uint32(data)
	head.PacketNo = binary.BigEndian.Uint16(data[4:6])
	head.ReplyNo = binary.BigEndian.Uint16(data[6:8])

	body := NewProtoMessage(head.MsgID)
	if body == nil {
		return nil, InvalidID
	}
	err := proto.Unmarshal(data[headSize:], body)
	if err != nil {
		return nil, err
	}
	head.Body = body
	return head, nil
}

func Marshal(msg proto.Message, packetNo, replyNo uint16) ([]byte, error) {
	msgId := GetMsgIDByName(string(msg.ProtoReflect().Type().Descriptor().FullName()))
	if msgId == 0 {
		return nil, InvalidMsg
	}
	body, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, headSize+len(body))
	binary.BigEndian.PutUint32(buf, msgId)
	binary.BigEndian.PutUint16(buf[4:], packetNo)
	binary.BigEndian.PutUint16(buf[6:], replyNo)
	copy(buf[headSize:], body)
	return buf, nil
}
