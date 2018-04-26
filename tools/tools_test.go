package tools

import (
	"math/big"
	"testing"

	"github.com/Baptist-Publication/chorus/eth/crypto"
	"github.com/Baptist-Publication/chorus/types"
)

func TestSig(t *testing.T) {
	privkey, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(privkey.PublicKey)

	tx := &types.BlockTx{
		GasLimit: big.NewInt(444),
		GasPrice: big.NewInt(444),
		Sender:   addr[:],
		Payload:  []byte("xxxx"),
	}

	if err := TxSign(tx, privkey); err != nil {
		t.Fatal(err)
	}

	bs, err := TxToBytes(tx)
	if err != nil {
		t.Fatal(err)
	}
	tx2 := &types.BlockTx{}
	if err = TxFromBytes(bs, tx2); err != nil {
		t.Fatal(err)
	}

	valid, err := TxVerifySignature(tx2)
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Fatal("valid error")
	}
}
