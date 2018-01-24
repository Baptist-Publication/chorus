package node

import (
	"errors"

	agtypes "github.com/Baptist-Publication/angine/types"
)

const (
	databaseCache   = 128
	databaseHandles = 1024
)

var (
	ErrRevertFromBackup = errors.New("revert from backup,not find data")
	ErrDataTransfer     = errors.New("data transfer err")
)

type AppTool struct {
	agtypes.BaseAppTool
	lastBlock LastBlockInfo
}

func (t *AppTool) Init(datadir string) error {
	if err := t.InitBaseApplication(BASE_APP_NAME, datadir); err != nil {
		return err
	}
	lb := NewLastBlockInfo()
	ret, err := t.LoadLastBlock(lb)
	if err != nil {
		return err
	}
	tmp, ok := ret.(*LastBlockInfo)
	if !ok {
		return ErrDataTransfer
	}
	t.lastBlock = *tmp
	return nil
}

func (t *AppTool) LastHeightHash() (agtypes.INT, []byte) {
	return agtypes.INT(t.lastBlock.Height), t.lastBlock.Hash
}

func (t *AppTool) BackupLastBlock(branchName string) error {
	return t.BackupLastBlockData(branchName, &t.lastBlock)
}

func (t *AppTool) SaveNewLastBlock(fromHeight agtypes.INT, fromAppHash []byte) error {
	newBranchBlock := LastBlockInfo{
		Height: fromHeight,
		Hash:   fromAppHash,
	}
	t.SaveLastBlock(newBranchBlock)
	// TODO
	return nil
}
