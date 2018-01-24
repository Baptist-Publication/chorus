package state

import (
	"fmt"
	"strings"

	cvtools "github.com/Baptist-Publication/chorus/src/tools"
	cvtypes "github.com/Baptist-Publication/chorus/src/types"
)

// for RemoteAccount, the key of saved KVdata should always like "<acc>-<user-key>"
type StateKvData struct {
	key  string
	data []byte
}

func (k *StateKvData) Init(key string, value []byte) {
	k.Reset(key, value)
}

func (k *StateKvData) Reset(key string, value []byte) {
	k.key = key
	k.data = value
}

func (k *StateKvData) Key() string {
	return k.key
}

func (k *StateKvData) Bytes() ([]byte, error) {
	return k.data, nil
}

func (k *StateKvData) Copy() cvtypes.StateDataItfc {
	cp := StateKvData{}
	cp.key = k.key
	cp.data = make([]byte, len(k.data))
	copy(cp.data, k.data)
	return &cp
}

func (k *StateKvData) OnCommit() error {
	return nil
}

func KvDataSplit(str string) (acc, key string) {
	if len(str) == 0 {
		return
	}
	index := strings.Index(str, "-")
	if index < 0 || index == len(str)-1 {
		return
	}
	return str[:index], str[index+1:]
}

// key is dangerous, so be carefull
func JointKvDataKey(acc, key string) (rkey string) {
	if len(acc) == 0 || len(key) == 0 {
		return
	}
	return fmt.Sprintf("%v-%v", acc, key)
}

//=============================FromBytesFunc===============================

func RemoteAccFromBytes(key string, data []byte) (cvtypes.StateDataItfc, error) {
	var acc RemoteAccount
	if err := cvtools.PbUnmarshal(data, &acc.RemoteAccountData); err != nil {
		return nil, err
	}
	return &acc, nil
}

func KVDataFromBytes(key string, data []byte) (cvtypes.StateDataItfc, error) {
	var kvd StateKvData
	kvd.Init(key, data)
	return &kvd, nil
}
