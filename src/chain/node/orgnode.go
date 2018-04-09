package node

import (
	"fmt"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/Baptist-Publication/angine"
	agtypes "github.com/Baptist-Publication/angine/types"
	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
)

const (
	RPCCollectSpecialVotes uint8 = iota
)

// OrgNode implements types.Core
type OrgNode struct {
	running int64

	logger *zap.Logger

	appname string

	Superior    Superior
	Angine      *angine.Angine
	AngineTune  *angine.Tunes
	Application agtypes.Application
	GenesisDoc  *agtypes.GenesisDoc
}

func (o *OrgNode) GetAppName() string {
	return o.appname
}

func (o *OrgNode) Start() error {
	if atomic.CompareAndSwapInt64(&o.running, 0, 1) {
		if err := o.Application.Start(); err != nil {
			o.Angine.Stop()
			return err
		}
		if err := o.Angine.Start(); err != nil {
			o.Application.Stop()
			o.Angine.Stop()
			return err
		}
		for o.Angine.Genesis() == nil {
			time.Sleep(500 * time.Millisecond)
		}
		return nil
	}
	return fmt.Errorf("already started")
}

func (o *OrgNode) Stop() bool {
	if atomic.CompareAndSwapInt64(&o.running, 1, 0) {
		o.Angine.Stop()
		o.Application.Stop()
		return true
	}
	return false
}

func (o *OrgNode) IsRunning() bool {
	return atomic.LoadInt64(&o.running) == 1
}

func (o *OrgNode) GetEngine() Engine {
	return o.Angine
}

func (o *OrgNode) GetPublicKey() (r crypto.PubKeyEd25519, b bool) {
	var pr *crypto.PubKeyEd25519
	pr, b = o.Angine.PrivValidator().GetPubKey().(*crypto.PubKeyEd25519)
	r = *pr
	return
}

func (o *OrgNode) GetPrivateKey() (r crypto.PrivKeyEd25519, b bool) {
	var pr *crypto.PrivKeyEd25519
	pr, b = o.Angine.PrivValidator().GetPrivKey().(*crypto.PrivKeyEd25519)
	r = *pr
	return
}

func (o *OrgNode) IsValidator() bool {
	_, vals := o.Angine.GetValidators()
	for i := range vals.Validators {
		if vals.Validators[i].GetPubKey().KeyString() == o.Angine.PrivValidator().GetPubKey().KeyString() {
			return true
		}
	}
	return false
}

func (o *OrgNode) GetChainID() string {
	return o.Angine.Genesis().ChainID
}

func (o *OrgNode) BroadcastTxSuperior(tx []byte) error {
	return o.Superior.BroadcastTx(tx)
}

func (o *OrgNode) PublishEvent(from string, block *agtypes.BlockCache, data []EventData, txhash []byte) error {
	return o.Superior.PublishEvent(from, block, data, txhash)
}

func (o *OrgNode) CodeExists(codehash []byte) bool {
	return o.Superior.CodeExists(codehash)
}
