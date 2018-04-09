package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/Baptist-Publication/angine/types"
	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	merkle "github.com/Baptist-Publication/chorus-module/lib/go-merkle"
	cl "github.com/Baptist-Publication/chorus-module/lib/go-rpc/client"
	"github.com/Baptist-Publication/chorus/src/chain/node"
	"github.com/Baptist-Publication/chorus/src/tools"
)

func transferRoutine(from crypto.PrivKeyEd25519, amount int) {
	sender := from.PubKey().(*crypto.PubKeyEd25519)[:]
	cli := newClient(defaultRPCServer)

	receiver := crypto.GenPrivKeyEd25519()
	to := receiver.PubKey().(*crypto.PubKeyEd25519)[:]

	nonce := queryNonce(cli, sender)

	for i := 0; i < amount; i++ {
		doTransfer(cli, from, to, nonce)
		nonce++
	}

	wg.Done()
}

func doTransfer(cli *cl.ClientJSONRPC, from crypto.PrivKeyEd25519, to []byte, nonce uint64) {
	sender := from.PubKey().(*crypto.PubKeyEd25519)[:]

	tx := &node.EcoTransferTx{}
	tx.PubKey = sender
	tx.To = to
	tx.Amount = big.NewInt(1)
	tx.Fee = big.NewInt(0)
	if nonce == 0 {
		tx.Nonce = queryNonce(cli, sender)
	} else {
		tx.Nonce = nonce
	}
	tx.TimeStamp = uint64(time.Now().UnixNano())
	tx.Signature, _ = tools.TxSign(tx, &from)
	txbytes, err := tools.TxToBytes(tx)
	panicErr(err)

	txbytes = types.WrapTx(node.EcoTransferTag, txbytes)

	_, err = jsonRPC(cli, txbytes)
	panicErr(err)

	hash := merkle.SimpleHashFromBinary(txbytes)
	fmt.Println("send ok :", hex.EncodeToString(hash))
}
