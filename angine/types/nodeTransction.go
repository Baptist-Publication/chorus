package types

import (
	"encoding/json"
)

const (
	CoinbaseFrom   = "0000000000000000000000000000000000000000000000000000000000000000"
	CoinbaseAmount = 100
)

type NodeTransaction struct {
	From      []byte `json:"from"`
	To        []byte `json:"to"`
	Signatrue []byte `json:"signatrue"`
	Amount    uint64 `json:"amount"`
}

func NewNodeTransaction(from, to []byte, amount uint64) *NodeTransaction {
	return &NodeTransaction{
		From:   from,
		To:     to,
		Amount: amount,
	}
}

func (nx *NodeTransaction) FillSign(sig []byte) {
	nx.Signatrue = sig
}

func (nx *NodeTransaction) ToBytes() []byte {
	bs, err := json.Marshal(nx)
	if err != nil {
		return nil
	}
	return bs
}

func (nx *NodeTransaction) FromBytes(data []byte) error {
	if err := json.Unmarshal(data, nx); err != nil {
		return err
	}

	return nil
}
