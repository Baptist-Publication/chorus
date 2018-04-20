package types

import (
	"bytes"
	"crypto/ecdsa"
	"math/big"

	"github.com/Baptist-Publication/chorus/src/eth/crypto"
	"github.com/Baptist-Publication/chorus/src/eth/rlp"
)

type EcoInitTokenTx struct {
	To     []byte   `json:"to"`
	Amount *big.Int `json:"amount"`
	Extra  []byte   `json:"extra"`
}

type EcoInitShareTx struct {
	To     []byte   `json:"to"`
	Amount *big.Int `json:"amount"`
	Extra  []byte   `json:"extra"`
}

type BlockTx struct {
	GasLimit  *big.Int
	GasPrice  *big.Int
	Nonce     uint64
	Sender    []byte
	Signature []byte
	Payload   []byte
}

func NewBlockTx(gasLimit, gasPrice *big.Int, nonce uint64, sender, payload []byte) *BlockTx {
	return &BlockTx{
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Nonce:    nonce,
		Sender:   sender,
		Payload:  payload,
	}
}

type TxEvmCommon struct {
	To     []byte
	Amount *big.Int
	Load   []byte
}

type TxShareTransfer struct {
	ShareSrc []byte
	ShareSig []byte
	ShareDst []byte
	Amount   *big.Int
}

func sigHash(tx *BlockTx) ([]byte, error) {
	txbytes, err := rlp.EncodeToBytes([]interface{}{
		tx.GasLimit,
		tx.GasPrice,
		tx.Nonce,
		tx.Sender,
		tx.Payload,
	})
	if err != nil {
		return nil, err
	}

	h := crypto.Sha256(txbytes)
	return h, nil
}

func (tx *BlockTx) Sign(privkey *ecdsa.PrivateKey) error {
	h, err := sigHash(tx)
	if err != nil {
		return err
	}

	sig, err := crypto.Sign(h, privkey)
	if err != nil {
		return err
	}

	tx.Signature = sig
	return nil
}

func (tx *BlockTx) VerifySignature() (bool, error) {
	if len(tx.Signature) == 0 {
		return false, nil
	}

	h, err := sigHash(tx)
	if err != nil {
		return false, err
	}

	pub, err := crypto.Ecrecover(h, tx.Signature)
	if err != nil {
		return false, err
	}
	addr := crypto.Keccak256(pub[1:])[12:]
	return bytes.Equal(tx.Sender, addr), nil
}

func (tx *BlockTx) Hash() []byte {
	bs, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil
	}

	return crypto.Sha256(bs)
}
