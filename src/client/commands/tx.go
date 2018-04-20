package commands

import (
	"fmt"
	"math/big"

	"github.com/Baptist-Publication/chorus/src/eth/common"
	ethtypes "github.com/Baptist-Publication/chorus/src/eth/core/types"
	"github.com/Baptist-Publication/chorus/src/eth/crypto"
	"github.com/Baptist-Publication/chorus/src/eth/rlp"
	"gopkg.in/urfave/cli.v1"

	"github.com/Baptist-Publication/angine/types"
	cl "github.com/Baptist-Publication/chorus-module/lib/go-rpc/client"
	"github.com/Baptist-Publication/chorus/src/chain/app"
	"github.com/Baptist-Publication/chorus/src/client/commons"
)

var (
	TxCommands = cli.Command{
		Name:     "tx",
		Usage:    "operations for transaction",
		Category: "Transaction",
		Subcommands: []cli.Command{
			{
				Name:   "send",
				Usage:  "send a transaction",
				Action: sendTx,
				Flags: []cli.Flag{
					anntoolFlags.payload,
					anntoolFlags.privkey,
					anntoolFlags.nonce,
					anntoolFlags.to,
					anntoolFlags.value,
				},
			},
		},
	}
)

func sendTx(ctx *cli.Context) error {

	privkey := ctx.String("privkey")

	var chainID string
	if !ctx.GlobalIsSet("target") {
		chainID = "chorus"
	} else {
		chainID = ctx.GlobalString("target")
	}

	nonce := ctx.Uint64("nonce")
	to := common.HexToAddress(ctx.String("to"))
	value := ctx.Int64("value")
	payload := ctx.String("payload")
	data := common.Hex2Bytes(payload)

	tx := ethtypes.NewTransaction(nonce, to, big.NewInt(value), big.NewInt(90000), big.NewInt(2), data)

	if privkey != "" {
		key, err := crypto.HexToECDSA(privkey)
		if err != nil {
			panic(err)
		}
		sig, _ := crypto.Sign(tx.SigHash(ethSigner).Bytes(), key)
		sigTx, _ := tx.WithSignature(ethSigner, sig)

		b, err := rlp.EncodeToBytes(sigTx)
		if err != nil {
			panic(err)
		}

		tmResult := new(types.RPCResult)
		clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
		_, err = clientJSON.Call("broadcast_tx_commit", []interface{}{chainID, types.WrapTx(app.EVMTxTag, b)}, tmResult)
		if err != nil {
			panic(err)
		}
		//res := (*tmResult).(*types.ResultBroadcastTxCommit)

		fmt.Println("tx result:", sigTx.Hash().Hex())
	}
	return nil
}
