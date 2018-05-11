package main

import (
	"fmt"
	"math/big"

	"time"

	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	"github.com/Baptist-Publication/chorus/eth/common"
	"github.com/Baptist-Publication/chorus/eth/crypto"
	ac "github.com/Baptist-Publication/chorus/module/lib/go-common"
	cl "github.com/Baptist-Publication/chorus/module/lib/go-rpc/client"
	"github.com/Baptist-Publication/chorus/tools"
	"github.com/Baptist-Publication/chorus/types"
)

func sendTx(privkey, toAddr string, value int64) error {
	sk, err := crypto.HexToECDSA(ac.SanitizeHex(privkey))
	panicErr(err)

	nonce := uint64(time.Now().UnixNano())

	btxbs, err := tools.ToBytes(&types.TxEvmCommon{
		To:     common.HexToAddress(toAddr).Bytes(),
		Amount: big.NewInt(value),
	})
	panicErr(err)

	tx := types.NewBlockTx(gasLimit, big.NewInt(0), nonce, crypto.PubkeyToAddress(sk.PublicKey).Bytes(), btxbs)
	tx.Signature, err = tools.SignSecp256k1(tx, crypto.FromECDSA(sk))
	panicErr(err)
	b, err := tools.ToBytes(tx)
	panicErr(err)

	tmResult := new(agtypes.RPCResult)
	clientJSON := cl.NewClientJSONRPC(logger, rpcTarget)
	_, err = clientJSON.Call("broadcast_tx_commit", []interface{}{append(types.TxTagAppEvmCommon, b...)}, tmResult)
	panicErr(err)

	res := (*tmResult).(*agtypes.ResultBroadcastTxCommit)
	fmt.Println("******************result", res)

	return nil
}
