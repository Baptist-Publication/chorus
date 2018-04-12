package commands

import (
	// "encoding/binary"
	"encoding/hex"
	//"encoding/json"
	"fmt"
	"math/big"
	//"strings"

	"gopkg.in/urfave/cli.v1"

	"github.com/Baptist-Publication/angine/types"
	ac "github.com/Baptist-Publication/chorus-module/lib/go-common"
	cl "github.com/Baptist-Publication/chorus-module/lib/go-rpc/client"
	//civil "github.com/Baptist-Publication/chorus/src/chain/node"
	"github.com/Baptist-Publication/chorus/src/client/commons"
	ethtypes "github.com/Baptist-Publication/chorus/src/eth/core/types"
	"github.com/Baptist-Publication/chorus/src/eth/rlp"
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
				Name:   "power",
				Usage:  "query account's vote power",
				Action: queryPower,
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
	var chainID string
	if !ctx.GlobalIsSet("target") {
		chainID = "chorus"
	} else {
		chainID = ctx.GlobalString("target")
	}
	nonce, err := getNonce(chainID, ctx.String("address"))
	if err != nil {
		return err
	}

	fmt.Println("query result:", nonce)

	return nil
}

func getNonce(chainID, addr string) (nonce uint64, err error) {
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)

	addrHex := ac.SanitizeHex(addr)
	adr, _ := hex.DecodeString(addrHex)
	query := append([]byte{1}, adr...)

	_, err = clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return 0, cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)
	//nonce = binary.LittleEndian.Uint64(res.Result.Data)
	rlp.DecodeBytes(res.Result.Data, &nonce)
	return nonce, nil
}

func queryBalance(ctx *cli.Context) error {
	var chainID string
	if !ctx.GlobalIsSet("target") {
		chainID = "chorus"
	} else {
		chainID = ctx.GlobalString("target")
	}
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)

	addrHex := ac.SanitizeHex(ctx.String("address"))
	addr, _ := hex.DecodeString(addrHex)
	query := append([]byte{2}, addr...)

	_, err := clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)

	balance := new(big.Int)
	rlp.DecodeBytes(res.Result.Data, balance)
	//balance := string(res.Result.Data)

	fmt.Println("query result:", balance)

	return nil
}

func queryShare(ctx *cli.Context) error {
	return nil
}

func queryPower(ctx *cli.Context) error {
	// var chainID string
	// if !ctx.GlobalIsSet("target") {
	// 	chainID = "chorus"
	// } else {
	// 	chainID = ctx.GlobalString("target")
	// }
	// clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	// tmResult := new(types.RPCResult)
	//
	// pubHex := ac.SanitizeHex(ctx.String("account_pubkey"))
	// pub, _ := hex.DecodeString(pubHex)
	// query := append([]byte{civil.QueryPower}, pub...)
	//
	// _, err := clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	// if err != nil {
	// 	return cli.NewExitError(err.Error(), 127)
	// }
	//
	// res := (*tmResult).(*types.ResultQuery)
	//
	// var info []string
	// if err := json.Unmarshal(res.Result.Data, &info); err != nil {
	// 	return cli.NewExitError(err.Error(), 127)
	// }
	//
	// fmt.Println("power:", info[0], "mheight:", info[1])

	return nil
}

func queryReceipt(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chainid is missing", 127)
	}
	chainID := ctx.GlobalString("target")
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)
	hashHex := ac.SanitizeHex(ctx.String("hash"))
	hash, _ := hex.DecodeString(hashHex)
	query := append([]byte{3}, hash...)
	_, err := clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)

	receiptdata := ethtypes.ReceiptForStorage{}
	rlp.DecodeBytes(res.Result.Data, &receiptdata)
	resultMap := map[string]interface{}{
		"code":              res.Result.Code,
		"txHash":            receiptdata.TxHash.Hex(),
		"contractAddress":   receiptdata.ContractAddress.Hex(),
		"cumulativeGasUsed": receiptdata.CumulativeGasUsed,
		"GasUsed":           receiptdata.GasUsed,
		"logs":              receiptdata.Logs,
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
