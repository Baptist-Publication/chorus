package types

import (
	"math/big"

	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus-module/xlib/def"
)

type ICivilTx interface {
	GetPubKey() []byte
	SetPubKey([]byte)
	PopSignature() []byte
	SetSignature([]byte)
	Sender(def.INT) []byte
	GetNonce() uint64
}

type CivilTx struct {
	sender []byte

	Nonce uint64   `json:"nonce"`
	Fee   *big.Int `json:"fee"`

	PubKey    []byte `json:"pubkey"`
	Signature []byte `json:"signature"`

	Extra     []byte `json:"extra"`
	TimeStamp uint64 `json:"timestamp"`
}

func (t *CivilTx) GetPubKey() []byte {
	return t.PubKey
}

func (t *CivilTx) SetPubKey(pk []byte) {
	t.PubKey = pk
}

func (t *CivilTx) PopSignature() []byte {
	s := t.Signature
	t.Signature = nil
	return s
}

func (t *CivilTx) SetSignature(s []byte) {
	t.Signature = s
}

func (t *CivilTx) Sender(height def.INT) []byte {
	if t.sender != nil || len(t.sender) > 0 {
		return t.sender
	}

	pubkey := crypto.PubKeyEd25519{}
	copy(pubkey[:], t.PubKey)
	t.sender = pubkey.Address(height)
	return t.sender
}

func (t *CivilTx) GetNonce() uint64 {
	return t.Nonce
}
