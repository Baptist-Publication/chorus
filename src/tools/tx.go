package tools

import (
	"bytes"
	"crypto/ecdsa"

	"github.com/Baptist-Publication/chorus/src/eth/crypto"
	"github.com/Baptist-Publication/chorus/src/eth/rlp"
	"github.com/Baptist-Publication/chorus/src/types"
)

func TxToBytes(tx interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(tx)
}

func TxFromBytes(bs []byte, tx interface{}) error {
	return rlp.DecodeBytes(bs, tx)
}

func sigHash(tx *types.BlockTx) ([]byte, error) {
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

func TxSign(tx *types.BlockTx, privkey *ecdsa.PrivateKey) error {
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

func TxVerifySignature(tx *types.BlockTx) (bool, error) {
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
