package p2p

import (
	"github.com/gogo/protobuf/proto"
)

//-----------------------------------------------------------------------------
// Messages

type P2plMsgItfc proto.Message

func UnmarshalP2pMsg(bz []byte) (P2plMsgItfc, error) {
	var ckMsg P2plMsgItfc
	err := proto.Unmarshal(bz, ckMsg)
	if err != nil {
		return nil, err
	}
	return ckMsg, err
}

func MarshalDataToMpMsg(msg proto.Message) []byte {
	var bs []byte
	var err error
	bs, err = proto.Marshal(msg)
	if err != nil {
		return nil
	}
	return bs
}
