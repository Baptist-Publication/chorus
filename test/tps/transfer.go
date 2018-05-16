package main

import (
	"errors"
	"fmt"
	"math/big"

	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	"github.com/Baptist-Publication/chorus/eth/common"
	"github.com/Baptist-Publication/chorus/eth/crypto"
	ac "github.com/Baptist-Publication/chorus/module/lib/go-common"
	cl "github.com/Baptist-Publication/chorus/module/lib/go-rpc/client"
	"github.com/Baptist-Publication/chorus/tools"
	"github.com/Baptist-Publication/chorus/types"
)

func send(client *cl.ClientJSONRPC, privkey, toAddr string, value int64, nonce uint64) error {
	sk, err := crypto.HexToECDSA(ac.SanitizeHex(privkey))
	panicErr(err)

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

	res := new(agtypes.ResultBroadcastTx)
	if client == nil {
		client = cl.NewClientJSONRPC(logger, rpcTarget)
	}
	_, err = client.Call("broadcast_tx_sync", []interface{}{append(types.TxTagAppEvmCommon, b...)}, res)
	panicErr(err)

	if res.Code != 0 {
		fmt.Println(res.Code, string(res.Data), res.Log)
		return errors.New(string(res.Data))
	}

	return nil
}
