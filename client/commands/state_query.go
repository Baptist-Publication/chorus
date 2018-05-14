package commands

import (
	"encoding/hex"
	"fmt"
	"math/big"

	pbtypes "github.com/Baptist-Publication/chorus/angine/protos/types"
	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	"github.com/Baptist-Publication/chorus/client/commons"
	ethtypes "github.com/Baptist-Publication/chorus/eth/core/types"
	"github.com/Baptist-Publication/chorus/eth/rlp"
	ac "github.com/Baptist-Publication/chorus/module/lib/go-common"
	cl "github.com/Baptist-Publication/chorus/module/lib/go-rpc/client"
	"github.com/Baptist-Publication/chorus/types"
	"gopkg.in/urfave/cli.v1"
	"errors"
)

var (
	QueryCommands = cli.Command{
		Name:     "query",
		Usage:    "operations for query state",
		Category: "Query",
		Subcommands: []cli.Command{
			{
				Name:   "nonce",
				Usage:  "query account's nonce",
				Action: queryNonce,
				Flags: []cli.Flag{
					anntoolFlags.addr,
				},
			},
			{
				Name:   "balance",
				Usage:  "query account's balance",
				Action: queryBalance,
				Flags: []cli.Flag{
					anntoolFlags.addr,
				},
			},
			{
				Name:   "share",
				Usage:  "query node's share",
				Action: queryShare,
				Flags: []cli.Flag{
					anntoolFlags.accountPubkey,
				},
			},
			{
				Name:   "receipt",
				Usage:  "",
				Action: queryReceipt,
				Flags: []cli.Flag{
					anntoolFlags.hash,
				},
			},
			{
				Name:   "isvalidator",
				Usage:  "query account is validator",
				Action: queryValidator,
				Flags: []cli.Flag{
					anntoolFlags.accountPubkey,
				},
			},
		},
	}
)

func queryNonce(ctx *cli.Context) error {
	nonce, err := getNonce(ctx.String("address"))
	if err != nil {
		return err
	}

	fmt.Println("query result:", nonce)

	return nil
}

func getNonce(addr string) (nonce uint64, err error) {
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(agtypes.ResultQuery)

	addrHex := ac.SanitizeHex(addr)
	adr, _ := hex.DecodeString(addrHex)
	query := append([]byte{types.QueryTypeNonce}, adr...)

	_, err = clientJSON.Call("query", []interface{}{query}, tmResult)
	if err != nil {
		return 0, cli.NewExitError(err.Error(), 127)
	}

	//nonce = binary.LittleEndian.Uint64(res.Result.Data)
	rlp.DecodeBytes(tmResult.Result.Data, &nonce)
	return nonce, nil
}

func queryBalance(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(agtypes.ResultQuery)

	addrHex := ac.SanitizeHex(ctx.String("address"))
	addr, _ := hex.DecodeString(addrHex)
	query := append([]byte{types.QueryTypeBalance}, addr...)

	_, err := clientJSON.Call("query", []interface{}{query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	balance := new(big.Int)
	rlp.DecodeBytes(tmResult.Result.Data, balance)
	if tmResult.Result.Code != pbtypes.CodeType_OK{
		fmt.Println("Error: ", tmResult.Result.Log)
		return errors.New(tmResult.Result.Log)
	}

	fmt.Println("query result:", balance)
	return nil
}

func queryShare(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(agtypes.ResultQuery)

	addrHex := ac.SanitizeHex(ctx.String("account_pubkey"))
	addr, _ := hex.DecodeString(addrHex)
	query := append([]byte{types.QueryTypeShare}, addr...)

	_, err := clientJSON.Call("query", []interface{}{query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	if tmResult.Result.IsErr() {
		return cli.NewExitError(tmResult.Result.Log, 127)
	}

	share := types.QueryShareResult{}
	rlp.DecodeBytes(tmResult.Result.Data, &share)

	fmt.Println("balance:", share.ShareBalance.String(), "guaranty:", share.ShareGuaranty.String(), "guaranty_height:", share.GHeight)
	return nil
}

func queryReceipt(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(agtypes.ResultQuery)
	hashHex := ac.SanitizeHex(ctx.String("hash"))
	hash, _ := hex.DecodeString(hashHex)
	query := append([]byte{types.QueryTypeReceipt}, hash...)
	_, err := clientJSON.Call("query", []interface{}{query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	receiptdata := ethtypes.ReceiptForStorage{}
	rlp.DecodeBytes(tmResult.Result.Data, &receiptdata)
	resultMap := map[string]interface{}{
		"code":              tmResult.Result.Code,
		"txHash":            receiptdata.TxHash.Hex(),
		"contractAddress":   receiptdata.ContractAddress.Hex(),
		"cumulativeGasUsed": receiptdata.CumulativeGasUsed,
		"GasUsed":           receiptdata.GasUsed,
		"logs":              receiptdata.Logs,
		"height":            receiptdata.Height,
	}
	fmt.Println("query result:", resultMap)

	return nil
}

func queryValidator(ctx *cli.Context) error {
	// var chainID string
	// if !ctx.GlobalIsSet("target") {
	// 	chainID = "chorus"
	// } else {
	// 	chainID = ctx.GlobalString("target")
	// }
	// clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	// tmResult := new(types.RPCResult)
	//
	// pubHex := strings.TrimPrefix(strings.ToUpper(ctx.String("account_pubkey")), "0x")
	//
	// _, err := clientJSON.Call("is_validator", []interface{}{chainID, pubHex}, tmResult)
	// if err != nil {
	// 	return cli.NewExitError(err.Error(), 127)
	// }
	//
	// res := (*tmResult).(*types.ResultQuery)
	//
	// fmt.Println("result:", res.Result.Data[0] == 1)
	return nil
}
