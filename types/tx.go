package types

import (
	"math/big"
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

type WorldRandTx struct {
	Height uint64
	Pubkey []byte
	Sig    []byte
}

type BlockTx struct {
	GasLimit  *big.Int
	GasPrice  *big.Int
	Nonce     uint64
	Sender    []byte
	Payload   []byte
	Signature []byte
}

type TxEvmCommon struct {
	To     []byte
	Amount *big.Int
	Load   []byte
}

type TxShareEco struct {
	Source    []byte
	Amount    *big.Int
	Signature []byte
}

type TxShareTransfer struct {
	ShareSrc []byte
	ShareDst []byte
	Amount   *big.Int
	ShareSig []byte
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

func (tx *BlockTx) SigObject() interface{} {
	return []interface{}{
		tx.GasLimit,
		tx.GasPrice,
		tx.Nonce,
		tx.Sender,
		tx.Payload,
	}
}

func (tx *TxShareEco) SigObject() interface{} {
	return []interface{}{
		tx.Source,
		tx.Amount,
	}
}

func (tx *TxShareTransfer) SigObject() interface{} {
	return []interface{}{
		tx.ShareSrc,
		tx.ShareDst,
		tx.Amount,
	}
}
