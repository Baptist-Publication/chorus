package tools

import "github.com/Baptist-Publication/chorus/eth/rlp"

func TxToBytes(tx interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(tx)
}

func TxFromBytes(bs []byte, tx interface{}) error {
	return rlp.DecodeBytes(bs, tx)
}
