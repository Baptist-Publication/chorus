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

type BlockTx struct {
	GasLimit  *big.Int
	GasPrice  *big.Int
	Nonce     uint64
	Sender    []byte
	Signature []byte
	Payload   []byte
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
