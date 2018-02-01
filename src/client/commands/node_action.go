package commands

import (
	"time"
	//"encoding/json"
	"fmt"

	"encoding/hex"
	"math/big"

	"gopkg.in/urfave/cli.v1"

	"github.com/Baptist-Publication/angine/types"
	//anginetypes "github.com/Baptist-Publication/angine/types"
	gcommon "github.com/Baptist-Publication/chorus-module/lib/go-common"
	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus-module/lib/go-merkle"
	"github.com/Baptist-Publication/chorus-module/lib/go-rpc/client"
	"github.com/Baptist-Publication/chorus/src/chain/node"
	"github.com/Baptist-Publication/chorus/src/client/commons"
	"github.com/Baptist-Publication/chorus/src/tools"
)

type nodeActions struct{}

var (
	nodeAction   = nodeActions{}
	NodeCommands = cli.Command{
		Name:     "tx",
		Usage:    "commands for node  operations",
		Category: "transaction",
		Subcommands: []cli.Command{
			{
				Name:   "transfer",
				Usage:  "transfer node balance from one node to another",
				Action: nodeAction.ChangeNodeBalance,
				Flags: []cli.Flag{
					anntoolFlags.privkey,
					anntoolFlags.peerPubkey,
					anntoolFlags.value,
					anntoolFlags.fee,
					anntoolFlags.nonce,
				},
			},
			{
				Name:   "mortgage",
				Usage:  "exchange node balance to voting power",
				Action: nodeAction.Mortgage,
				Flags: []cli.Flag{
					anntoolFlags.privkey,
					anntoolFlags.value,
					anntoolFlags.fee,
					anntoolFlags.nonce,
				},
			},
			{
				Name:   "redemption",
				Usage:  "exchange node voting power to balance",
				Action: nodeAction.Redemption,
				Flags: []cli.Flag{
					anntoolFlags.privkey,
					anntoolFlags.value,
					anntoolFlags.fee,
					anntoolFlags.nonce,
				},
			},
		},
	}
)

func (act nodeActions) jsonRPC(chainID string, p []byte) (*types.RPCResult, error) {
	clt := rpcclient.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)
	_, err := clt.Call("broadcast_tx_sync", []interface{}{chainID, p}, tmResult)
	if err != nil {
		return nil, err
	}
	return tmResult, nil
}

func (act nodeActions) ChangeNodeBalance(ctx *cli.Context) error {
	var chainID string
	if !ctx.GlobalIsSet("target") {
		chainID = "chorus"
	} else {
		chainID = ctx.GlobalString("target")
	}
	if !ctx.IsSet("privkey") || !ctx.IsSet("peer_pubkey") || !ctx.IsSet("value") {
		return cli.NewExitError("from node privkey , peer pubkey and change value is required", 127)
	}

	priv, to := gcommon.SanitizeHex(ctx.String("privkey")), gcommon.SanitizeHex(ctx.String("peer_pubkey"))

	privBytes, err := hex.DecodeString(priv)
	if err != nil {
		return err
	}
	privKey := crypto.PrivKeyEd25519{}
	copy(privKey[:], privBytes)
	frompubkey := privKey.PubKey().(*crypto.PubKeyEd25519)

	topubkey32, err := types.StringTo32byte(to)
	if err != nil {
		return err
	}
	topubkey := crypto.PubKeyEd25519{}
	copy(topubkey[:], topubkey32[:])

	value := ctx.Int64("value")

	tx := &node.EcoTransferTx{}
	tx.PubKey = frompubkey[:]
	tx.To = topubkey[:]
	tx.Amount = big.NewInt(value)
	if ctx.IsSet("nonce") {
		tx.Nonce = ctx.Uint64("nonce")
	} else {
		tx.Nonce, err = getNonce(chainID, frompubkey.KeyString())
	}
	tx.TimeStamp = uint64(time.Now().UnixNano())
	if ctx.IsSet("fee") {
		tx.Fee = big.NewInt(ctx.Int64("fee"))
	} else {
		tx.Fee = big.NewInt(0)
	}

	tx.Signature, _ = tools.TxSign(tx, &privKey)
	txbytes, err := tools.TxToBytes(tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	txbytes = types.WrapTx(node.EcoTransferTag, txbytes)

	_, err = act.jsonRPC(chainID, txbytes)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	//hash, _ := tools.TxHash(tx)
	hash := merkle.SimpleHashFromBinary(txbytes)
	fmt.Println("send ok :", hex.EncodeToString(hash))

	return nil
}

func (act nodeActions) Mortgage(ctx *cli.Context) error {
	var chainID string
	if !ctx.GlobalIsSet("target") {
		chainID = "chorus"
	} else {
		chainID = ctx.GlobalString("target")
	}
	if !ctx.IsSet("privkey") || !ctx.IsSet("value") {
		return cli.NewExitError("from node privkey ,change value is required", 127)
	}

	priv := gcommon.SanitizeHex(ctx.String("privkey"))

	privBytes, err := hex.DecodeString(priv)
	if err != nil {
		return err
	}
	privKey := crypto.PrivKeyEd25519{}
	copy(privKey[:], privBytes)
	frompubkey := privKey.PubKey().(*crypto.PubKeyEd25519)

	value := ctx.Int64("value")

	tx := &node.EcoMortgageTx{}
	tx.PubKey = frompubkey[:]
	tx.Amount = big.NewInt(value)
	if ctx.IsSet("nonce") {
		tx.Nonce = ctx.Uint64("nonce")
	} else {
		tx.Nonce, err = getNonce(chainID, frompubkey.KeyString())
	}
	tx.TimeStamp = uint64(time.Now().UnixNano())
	if ctx.IsSet("fee") {
		tx.Fee = big.NewInt(ctx.Int64("fee"))
	} else {
		tx.Fee = big.NewInt(0)
	}

	tx.Signature, _ = tools.TxSign(tx, &privKey)
	txbytes, err := tools.TxToBytes(tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	txbytes = types.WrapTx(node.EcoMortgageTag, txbytes)

	_, err = act.jsonRPC(chainID, txbytes)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	//hash, _ := tools.TxHash(tx)
	hash := merkle.SimpleHashFromBinary(txbytes)
	fmt.Println("send ok :", hex.EncodeToString(hash))

	return nil
}

func (act nodeActions) Redemption(ctx *cli.Context) error {
	var chainID string
	if !ctx.GlobalIsSet("target") {
		chainID = "chorus"
	} else {
		chainID = ctx.GlobalString("target")
	}
	if !ctx.IsSet("privkey") || !ctx.IsSet("value") {
		return cli.NewExitError("from node privkey , change value is required", 127)
	}

	priv := gcommon.SanitizeHex(ctx.String("privkey"))

	privBytes, err := hex.DecodeString(priv)
	if err != nil {
		return err
	}
	privKey := crypto.PrivKeyEd25519{}
	copy(privKey[:], privBytes)
	frompubkey := privKey.PubKey().(*crypto.PubKeyEd25519)

	value := ctx.Int64("value")

	tx := &node.EcoRedemptionTx{}
	tx.PubKey = frompubkey[:]
	tx.Amount = big.NewInt(value)
	if ctx.IsSet("nonce") {
		tx.Nonce = ctx.Uint64("nonce")
	} else {
		tx.Nonce, err = getNonce(chainID, frompubkey.KeyString())
	}
	tx.TimeStamp = uint64(time.Now().UnixNano())
	if ctx.IsSet("fee") {
		tx.Fee = big.NewInt(ctx.Int64("fee"))
	} else {
		tx.Fee = big.NewInt(0)
	}

	tx.Signature, _ = tools.TxSign(tx, &privKey)
	txbytes, err := tools.TxToBytes(tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	txbytes = types.WrapTx(node.EcoRedemptionTag, txbytes)

	_, err = act.jsonRPC(chainID, txbytes)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	//hash, _ := tools.TxHash(tx)
	hash := merkle.SimpleHashFromBinary(txbytes)
	fmt.Println("send ok :", hex.EncodeToString(hash))

	return nil
}
