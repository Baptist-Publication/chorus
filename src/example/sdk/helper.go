package sdk

import (
	"bytes"
	"fmt"

	"github.com/golang/protobuf/proto"
	remote "github.com/Baptist-Publication/chorus/src/chain/app/remote"
)

// TODO some inputs (such as length,nil...etc) safe check

type Helper struct {
	do remote.RemoteAppService_ExecuteServer
}

func (h *Helper) Init(do remote.RemoteAppService_ExecuteServer) {
	h.do = do
}

func (h *Helper) AddData(k string, v []byte) error {
	prefix := []byte("add_data") // TODO Random Gen as a serial No.
	var cmd remote.S_AddDataCmd
	//cmd.Key, _ = tools.HashKeccak(k)
	//cmd.Value, _ = tools.HashKeccak(v)
	cmd.Key, cmd.Value = k, v
	send := GenCmd(prefix, remote.S_CmdType_AddData, ToBytes(&cmd))
	var err error
	if err = h.do.Send(&send); err != nil {
		return fmt.Errorf("send add data err:%v", err)
	}
	var rev *remote.ClientCmd
	if rev, err = h.do.Recv(); err != nil {
		return fmt.Errorf("send add data err:%v", err)
	}
	if !bytes.Equal(rev.Hash, prefix) || rev.Type != remote.C_CmdType_AddDataRes {
		return fmt.Errorf("cmd mismatch,recv:%v,revType:%v", rev.Hash, rev.Type)
	}
	var ret remote.C_AddDataCmd
	ToMsg(rev.Data, &ret)
	return checkRetResult(ret.GetResult())
}

func (h *Helper) QueryData(k string) ([]byte, error) {
	prefix := []byte("get_data") // TODO Random Gen as a serial No.
	var cmd remote.S_QueryDataCmd
	//cmd.Key, _ = tools.HashKeccak(k)
	cmd.Key = k
	send := GenCmd(prefix, remote.S_CmdType_QueryData, ToBytes(&cmd))
	var err error
	if err = h.do.Send(&send); err != nil {
		return nil, fmt.Errorf("send get data err:%v", err)
	}
	var rev *remote.ClientCmd
	if rev, err = h.do.Recv(); err != nil {
		return nil, fmt.Errorf("send get data err:%v", err)
	}
	if !bytes.Equal(rev.Hash, prefix) || rev.Type != remote.C_CmdType_QueryDataRes {
		return nil, fmt.Errorf("cmd error,recv:", rev.Hash, ",revType:", rev.Type)
	}
	var ret remote.C_QueryDataCmd
	ToMsg(rev.Data, &ret)
	err = checkRetResult(ret.GetResult())
	return ret.Value, err
}

func checkRetResult(ret *remote.PbResult) error {
	code := ret.GetCode()
	if code != remote.CodeType_OK {
		return fmt.Errorf("ret code is:%v,log:%v", code, ret.GetLog())
	}
	return nil
}

func GenCmd(hash []byte, tp remote.S_CmdType, data []byte) (cmd remote.ServerCmd) {
	cmd.Hash = hash
	cmd.Type = tp
	cmd.Data = data
	return
}

func ToBytes(msg proto.Message) []byte {
	ret, err := proto.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return ret
}

func ToMsg(data []byte, msg proto.Message) error {
	err := proto.Unmarshal(data, msg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
