package p2p

import (
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"
)

//-----------------------------------------------------------------------------
// Messages

type P2pMsgItfc proto.Message

func UnmarshalP2pMsg(bz []byte) (P2pMsgItfc, error) {
	var ckMsg CheckMessage
	err := proto.Unmarshal(bz, &ckMsg)
	if err != nil {
		return nil, err
	}
	var msgItfc P2pMsgItfc
	msgItfc, err = MsgFromType(ckMsg.GetType())
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(ckMsg.GetData(), msgItfc)
	return msgItfc, err
}

func MsgFromType(typ MsgType) (msgItfc P2pMsgItfc, err error) {
	switch typ {
	case MsgType_BlockPart:
		msgItfc = &BlockPartMessage{}
	default:
		return nil, errors.New(fmt.Sprintf("unmarshal,unknown p2p msg type:%v", typ))
	}
	return
}

func GetMessageType(msg proto.Message) MsgType {
	switch msg.(type) {
	case *BlockPartMessage:
		return MsgType_BlockPart
	}
	return MsgType_None
}

func MarshalDataToChkMsg(msg proto.Message) []byte {
	msgType := GetMessageType(msg)
	if msgType == MsgType_None {
		return nil
	}
	var bs, finbs []byte
	var err error
	bs, err = proto.Marshal(msg)
	if err != nil {
		return nil
	}
	msgBase := &CheckMessage{
		Type: msgType,
		Data: bs,
	}
	finbs, err = proto.Marshal(msgBase)
	if err != nil {
		return nil
	}
	return finbs
}

// Must ok
func MarshalData(msg P2pMsgItfc) []byte {
	bs, _ := proto.Marshal(msg)
	return bs
}

func (m *BlockPartMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[BlockPartMessage %v]", m)
}
