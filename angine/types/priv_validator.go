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

package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	pbtypes "github.com/Baptist-Publication/chorus/angine/protos/types"
	"github.com/Baptist-Publication/chorus/module/lib/ed25519"
	. "github.com/Baptist-Publication/chorus/module/lib/go-common"
	"github.com/Baptist-Publication/chorus/module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus/module/xlib/def"
	"go.uber.org/zap"
)

const (
	stepNone      = 0 // Used to distinguish the initial state
	stepPropose   = 1
	stepPrevote   = 2
	stepPrecommit = 3
)

func voteToStep(vote *pbtypes.Vote) int8 {
	switch vote.Data.Type {
	case pbtypes.VoteType_Prevote:
		return stepPrevote
	case pbtypes.VoteType_Precommit:
		return stepPrecommit
	default:
		PanicSanity("Unknown vote type")
		return 0
	}
}

type PrivValidatorJSON struct {
	PubKey        crypto.StPubKey    `json:"pub_key"`
	Coinbase      Bytes              `json:"coin_base"`
	LastHeight    def.INT            `json:"last_height"`
	LastRound     def.INT            `json:"last_round"`
	LastStep      int8               `json:"last_step"`
	LastSignature crypto.StSignature `json:"last_signature"` // so we dont lose signatures
	LastSignBytes Bytes              `json:"last_signbytes"` // so we dont lose signatures

	// PrivKey should be empty if a Signer other than the default is being used.
	PrivKey crypto.StPrivKey `json:"priv_key"`
}

type PrivValidator struct {
	PrivValidatorJSON
	Signer `json:"-"`

	address Bytes

	// For persistence.
	// Overloaded for testing.
	filePath string
	mtx      sync.Mutex

	logger *zap.Logger
}

func (pv *PrivValidator) CopyReset() (cp PrivValidator) {
	cp.PubKey = pv.PubKey
	cp.Coinbase = pv.Coinbase
	cp.PrivKey = pv.PrivKey
	cp.Signer = NewDefaultSigner(pv.GetPrivKey())
	return
}

func (pv *PrivValidator) GetPubKey() crypto.PubKey {
	return pv.PubKey.PubKey
}

// func (pv *PrivValidator) CoinbaseKey() crypto.PubKey {
// 	return pv.Coinbase.PubKey
// }

func (pv *PrivValidator) GetLastSignature() crypto.Signature {
	return pv.LastSignature.Signature
}

func (pv *PrivValidator) GetPrivKey() crypto.PrivKey {
	return pv.PrivKey.PrivKey
}

func (pv *PrivValidator) MarshalJSON() ([]byte, error) {
	return json.Marshal(&pv.PrivValidatorJSON)
}

func (pv *PrivValidator) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &pv.PrivValidatorJSON); err != nil {
		return err
	}
	return nil
}

// This is used to sign votes.
// It is the caller's duty to verify the msg before calling Sign,
// eg. to avoid double signing.
// Currently, the only callers are SignVote and SignProposal
type Signer interface {
	Sign(msg []byte) crypto.Signature
}

// Implements Signer
type DefaultSigner struct {
	priv crypto.PrivKey
}

func NewDefaultSigner(priv crypto.PrivKey) *DefaultSigner {
	return &DefaultSigner{priv: priv}
}

// Implements Signer
func (ds *DefaultSigner) Sign(msg []byte) crypto.Signature {
	return ds.priv.Sign(msg)
}

func (privVal *PrivValidator) SetSigner(s Signer) {
	privVal.Signer = s
}

func (privVal *PrivValidator) GetCoinbase() []byte {
	return privVal.Coinbase[:]
}

// Generates a new validator with private key.
func GenPrivValidator(logger *zap.Logger) *PrivValidator {
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], crypto.CRandBytes(32))
	pubKeyBytes := ed25519.MakePublicKey(privKeyBytes)
	pubKey := crypto.PubKeyEd25519(*pubKeyBytes)
	privKey := crypto.PrivKeyEd25519(*privKeyBytes)

	return &PrivValidator{
		PrivValidatorJSON: PrivValidatorJSON{
			PubKey:        crypto.StPubKey{&pubKey},
			PrivKey:       crypto.StPrivKey{&privKey},
			LastHeight:    0,
			LastRound:     0,
			LastStep:      stepNone,
			LastSignature: crypto.StSignature{nil},
			LastSignBytes: nil,
		},
		filePath: "",
		Signer:   NewDefaultSigner(&privKey),
		logger:   logger,
	}
}

func LoadPrivValidator(logger *zap.Logger, filePath string) *PrivValidator {
	privValJSONBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		Exit(err.Error())
	}
	var privVal PrivValidator
	err = json.Unmarshal(privValJSONBytes, &privVal)
	if err != nil {
		Exit(Fmt("Error reading PrivValidator from %v: %v\n", filePath, err))
	}
	privVal.filePath = filePath
	privVal.Signer = NewDefaultSigner(privVal.PrivKey.PrivKey)
	privVal.logger = logger
	return &privVal
}

func LoadOrGenPrivValidator(logger *zap.Logger, filePath string) *PrivValidator {
	var privValidator *PrivValidator
	if _, err := os.Stat(filePath); err == nil {
		privValidator = LoadPrivValidator(logger, filePath)
		if logger != nil {
			logger.Sugar().Infow("Loaded PrivValidator", "file", filePath, "privValidator", privValidator)
		}
	} else {
		privValidator = GenPrivValidator(logger)
		privValidator.SetFile(filePath)
		privValidator.Save()
		if logger != nil {
			logger.Info("Generated PrivValidator", zap.String("file", filePath))
		}
	}
	return privValidator
}

func (privVal *PrivValidator) SetFile(filePath string) {
	privVal.mtx.Lock()
	defer privVal.mtx.Unlock()
	privVal.filePath = filePath
}

func (privVal *PrivValidator) Save() {
	privVal.mtx.Lock()
	defer privVal.mtx.Unlock()
	privVal.save()
}

func (privVal *PrivValidator) save() {
	if privVal.filePath == "" {
		PanicSanity("Cannot save PrivValidator: filePath not set")
	}
	jsonBytes, err := json.MarshalIndent(privVal, "", "\t")
	if err != nil {
		PanicCrisis(fmt.Sprintf("json marshal indent err:", err))
	}
	err = WriteFileAtomic(privVal.filePath, jsonBytes, 0600)
	if err != nil {
		// `@; BOOM!!!
		PanicCrisis(err)
	}
}

// NOTE: Unsafe!
func (privVal *PrivValidator) Reset() {
	privVal.LastHeight = 0
	privVal.LastRound = 0
	privVal.LastStep = 0
	privVal.LastSignature = crypto.StSignature{nil}
	privVal.LastSignBytes = nil
	privVal.Save()
}

func (privVal *PrivValidator) GetAddress() []byte {
	if len(privVal.address) == 0 {
		addr := privVal.PubKey.Address()
		privVal.address = addr
		return addr
	}
	return privVal.address
}

func (privVal *PrivValidator) SignVote(chainID string, vote *pbtypes.Vote) error {
	privVal.mtx.Lock()
	defer privVal.mtx.Unlock()
	vdata := vote.GetData()
	signature, err := privVal.signBytesHRS(vdata.Height, vdata.Round, voteToStep(vote), SignBytes(chainID, vdata))
	if err != nil {
		return errors.New(Fmt("Error signing vote: %v", err))
	}
	vote.Signature = signature.Bytes()
	return nil
}

func (privVal *PrivValidator) SignProposal(chainID string, proposal *pbtypes.Proposal) error {
	privVal.mtx.Lock()
	defer privVal.mtx.Unlock()
	pdata := proposal.GetData()
	signature, err := privVal.signBytesHRS(pdata.Height, pdata.Round, stepPropose, SignBytes(chainID, pdata))
	if err != nil {
		return errors.New(Fmt("Error signing proposal: %v", err))
	}
	proposal.Signature = signature.Bytes()
	return nil
}

// check if there's a regression. Else sign and write the hrs+signature to disk
func (privVal *PrivValidator) signBytesHRS(height, round def.INT, step int8, signBytes []byte) (crypto.Signature, error) {
	// If height regression, err
	if privVal.LastHeight > height {
		return nil, errors.New("Height regression")
	}
	// More cases for when the height matches
	if privVal.LastHeight == height {
		// If round regression, err
		if privVal.LastRound > round {
			return nil, errors.New("Round regression")
		}
		// If step regression, err
		if privVal.LastRound == round {
			if privVal.LastStep > step {
				return nil, errors.New("Step regression")
			} else if privVal.LastStep == step {
				if privVal.LastSignBytes != nil {
					if privVal.LastSignature.IsZero() {
						PanicSanity("privVal: LastSignature is nil but LastSignBytes is not!")
					}
					// so we dont sign a conflicting vote or proposal
					// NOTE: proposals are non-deterministic (include time),
					// so we can actually lose them, but will still never sign conflicting ones
					if bytes.Equal(privVal.LastSignBytes, signBytes) {
						if privVal.logger != nil {
							privVal.logger.Sugar().Infof("Using privVal.LastSignature: %X", privVal.LastSignature)
						}
						return privVal.LastSignature.Signature, nil
					}
				}
				return nil, errors.New("Step regression")
			}
		}
	}

	// Sign
	signature := privVal.Sign(signBytes)

	// Persist height/round/step
	privVal.LastHeight = height
	privVal.LastRound = round
	privVal.LastStep = step
	privVal.LastSignature = crypto.StSignature{signature}
	privVal.LastSignBytes = signBytes
	privVal.save()

	return signature, nil

}

func (privVal *PrivValidator) String() string {
	return fmt.Sprintf("PrivValidator{%X LH:%v, LR:%v, LS:%v}", privVal.GetAddress(), privVal.LastHeight, privVal.LastRound, privVal.LastStep)
}

//-------------------------------------

type PrivValidatorsByAddress []*PrivValidator

func (pvs PrivValidatorsByAddress) Len() int {
	return len(pvs)
}

func (pvs PrivValidatorsByAddress) Less(i, j int) bool {
	return bytes.Compare(pvs[i].GetAddress(), pvs[j].GetAddress()) == -1
}

func (pvs PrivValidatorsByAddress) Swap(i, j int) {
	it := pvs[i]
	pvs[i] = pvs[j]
	pvs[j] = it
}
