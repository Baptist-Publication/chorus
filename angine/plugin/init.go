// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

import (
	"go.uber.org/zap"

	"github.com/Baptist-Publication/chorus/angine/blockchain/refuse_list"
	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	"github.com/Baptist-Publication/chorus/module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus/module/lib/go-db"
	"github.com/Baptist-Publication/chorus/module/lib/go-p2p"
)

const (
	PluginNoncePrefix = "pn-"
)

type (
	InitParams struct {
		Logger     *zap.Logger
		DB         db.DB
		Switch     *p2p.Switch
		PrivKey    crypto.PrivKeyEd25519
		RefuseList *refuse_list.RefuseList
		Validators **agtypes.ValidatorSet
	}

	ReloadParams struct {
		Logger     *zap.Logger
		DB         db.DB
		Switch     *p2p.Switch
		PrivKey    crypto.PrivKeyEd25519
		RefuseList *refuse_list.RefuseList
		Validators **agtypes.ValidatorSet
	}

	BeginBlockParams struct {
		Block *agtypes.BlockCache
	}

	BeginBlockReturns struct {
	}

	ExecBlockParams struct {
		Block       *agtypes.BlockCache
		EventSwitch agtypes.EventSwitch
		EventCache  agtypes.EventCache
		ValidTxs    agtypes.Txs
		InvalidTxs  []agtypes.ExecuteInvalidTx
	}

	ExecBlockReturns struct {
	}

	EndBlockParams struct {
		Block *agtypes.BlockCache
	}

	EndBlockReturns struct {
		NextValidatorSet *agtypes.ValidatorSet
	}

	// IPlugin defines the behavior of the core plugins
	IPlugin interface {
		agtypes.Eventable

		// DeliverTx return false means the tx won't be pass on to proxy app
		DeliverTx(tx []byte, i int) (bool, error)

		// CheckTx return false means the tx won't be pass on to proxy app
		CheckTx(tx []byte) (bool, error)

		// BeginBlock just mock the abci Blockaware interface
		BeginBlock(*BeginBlockParams) (*BeginBlockReturns, error)

		// ExecBlock receives block
		ExecBlock(*ExecBlockParams) (*ExecBlockReturns, error)

		// EndBlock just mock the abci Blockaware interface
		EndBlock(*EndBlockParams) (*EndBlockReturns, error)

		// Reset is called when u don't need to maintain the plugin status
		Reset()

		// InitPlugin custom the initialization of the plugin
		Init(*InitParams)

		// Reload reloads private fields of the plugin
		Reload(*ReloadParams)

		Stop()
	}
)
