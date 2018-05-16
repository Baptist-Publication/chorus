package commands

import (
	_ "encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"reflect"
	"strings"

	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	"github.com/Baptist-Publication/chorus/client/commons"
	"github.com/Baptist-Publication/chorus/eth/accounts/abi"
	"github.com/Baptist-Publication/chorus/eth/common"
	ethtypes "github.com/Baptist-Publication/chorus/eth/core/types"
	"github.com/Baptist-Publication/chorus/eth/crypto"
	"github.com/Baptist-Publication/chorus/eth/rlp"
	ac "github.com/Baptist-Publication/chorus/module/lib/go-common"
	cl "github.com/Baptist-Publication/chorus/module/lib/go-rpc/client"
	"github.com/Baptist-Publication/chorus/tools"
	"github.com/Baptist-Publication/chorus/types"
	"github.com/bitly/go-simplejson"
	"gopkg.in/urfave/cli.v1"
)

var (
	// gasLimit is the amount we will borrow from gas pool
	gasLimit = big.NewInt(210000)

	gasPrice = big.NewInt(2)
	//ContractCommands defines a more git-like subcommand system
	EVMCommands = cli.Command{
		Name:     "evm",
		Usage:    "operations for evm contract",
		Category: "Contract",
		Subcommands: []cli.Command{
			{
				Name:   "create",
				Usage:  "create a new contract",
				Action: createContract,
				Flags: []cli.Flag{
					anntoolFlags.addr,
					anntoolFlags.bytecode,
					anntoolFlags.privkey,
					anntoolFlags.nonce,
					anntoolFlags.abif,
					anntoolFlags.callf,
				},
			},
			{
				Name:   "execute",
				Usage:  "execute a new contract",
				Action: executeContract,
				Flags: []cli.Flag{
					anntoolFlags.addr,
					anntoolFlags.payload,
					anntoolFlags.privkey,
					anntoolFlags.nonce,
					anntoolFlags.abif,
					anntoolFlags.callf,
				},
			},
			{
				Name:   "read",
				Usage:  "read a contract",
				Action: readContract,
				Flags: []cli.Flag{
					anntoolFlags.payload,
					anntoolFlags.privkey,
					anntoolFlags.nonce,
					anntoolFlags.abistr,
					anntoolFlags.callstr,
					anntoolFlags.to,
					anntoolFlags.abif,
					anntoolFlags.callf,
				},
			},
			{
				Name:   "querycontract",
				Usage:  "read a contract",
				Action: queryContract,
				Flags: []cli.Flag{
					anntoolFlags.payload,
					anntoolFlags.abistr,
					anntoolFlags.callstr,
					anntoolFlags.to,
					anntoolFlags.abif,
					anntoolFlags.callf,
				},
			},
			{
				Name:   "exist",
				Usage:  "check if a contract has been deployed",
				Action: existContract,
				Flags: []cli.Flag{
					anntoolFlags.callf,
				},
			},
		},
	}
)

func readContract(ctx *cli.Context) error {
	json, err := getCallParamsJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	aabbii, err := getAbiJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	function := json.Get("function").MustString()
	if !aabbii.Methods[function].Const {
		fmt.Printf("we can only read constant method, %s is not! Any consequence is on you.\n", function)
	}
	params := json.Get("params").MustArray()
	contractAddress := ac.SanitizeHex(json.Get("contract").MustString())
	args, err := commons.ParseArgs(function, *aabbii, params)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	data, err := aabbii.Pack(function, args...)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	privkey := ctx.String("privkey")
	nonce := ctx.Uint64("nonce")
	to := common.HexToAddress(contractAddress)

	if privkey == "" {
		privkey = json.Get("privkey").MustString()
	}
	if privkey == "" {
		panic("should provide privkey.")
	}
	key, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	from := crypto.PubkeyToAddress(key.PublicKey)
	privkey = ac.SanitizeHex(privkey)
	tx := ethtypes.NewTransaction(nonce, from, to, big.NewInt(0), gasLimit, gasPrice, data)

	b, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	query := append([]byte{types.QueryTypeContract}, b...)
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	res := new(agtypes.ResultQuery)
	_, err = clientJSON.Call("query", []interface{}{query}, res)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	hex := common.Bytes2Hex(res.Result.Data)
	fmt.Println("query result:", hex)
	parseResult, _ := unpackResult(function, *aabbii, string(res.Result.Data))
	fmt.Println("parse result:", reflect.TypeOf(parseResult), parseResult)

	return nil
}

func unpackResult(method string, abiDef abi.ABI, output string) (interface{}, error) {
	m, ok := abiDef.Methods[method]
	if !ok {
		return nil, errors.New("No such method")
	}
	if len(m.Outputs) == 0 {
		return nil, errors.New("method " + m.Name + " doesn't have any returns")
	}
	if len(m.Outputs) == 1 {
		var result interface{}
		parsedData := common.ParseData(output)
		if err := abiDef.Unpack(&result, method, parsedData); err != nil {
			fmt.Println("error:", err)
			return nil, err
		}
		if strings.Index(m.Outputs[0].Type.String(), "bytes") == 0 {
			b := result.([]byte)
			idx := 0
			for i := 0; i < len(b); i++ {
				if b[i] == 0 {
					idx = i
				} else {
					break
				}
			}
			b = b[idx+1:]
			return fmt.Sprintf("%s", b), nil
		}
		return result, nil
	}
	d := common.ParseData(output)
	var result []interface{}
	if err := abiDef.Unpack(&result, method, d); err != nil {
		fmt.Println("fail to unpack outpus:", err)
		return nil, err
	}

	retVal := map[string]interface{}{}
	for i, output := range m.Outputs {
		if strings.Index(output.Type.String(), "bytes") == 0 {
			b := result[i].([]byte)
			idx := 0
			for i := 0; i < len(b); i++ {
				if b[i] == 0 {
					idx = i
				} else {
					break
				}
			}
			b = b[idx+1:]
			retVal[output.Name] = fmt.Sprintf("%s", b)
		} else {
			retVal[output.Name] = result[i]
		}
	}
	return retVal, nil
}

func executeContract(ctx *cli.Context) error {
	json, err := getCallParamsJSON(ctx)
	if err != nil {
		return err
	}
	aabbii, err := getAbiJSON(ctx)
	if err != nil {
		return err
	}
	function := json.Get("function").MustString()
	params := json.Get("params").MustArray()
	contractAddress := ac.SanitizeHex(json.Get("contract").MustString())
	args, err := commons.ParseArgs(function, *aabbii, params)
	if err != nil {
		panic(err)
	}

	data, err := aabbii.Pack(function, args...)
	if err != nil {
		return err
	}

	privkey := ctx.String("privkey")
	if privkey == "" {
		privkey = json.Get("privkey").MustString()
	}
	if privkey == "" {
		return cli.NewExitError("privkey is required", 123)
	}
	key, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return cli.NewExitError(err.Error(), 123)
	}
	from := crypto.PubkeyToAddress(key.PublicKey)
	to := common.HexToAddress(contractAddress)
	nonce := ctx.Uint64("nonce")

	bodyTx := &types.TxEvmCommon{
		To:   to[:],
		Load: data,
	}
	bodyBs, err := tools.ToBytes(bodyTx)
	if err != nil {
		return cli.NewExitError(err.Error(), 123)
	}

	tx := types.NewBlockTx(gasLimit, gasPrice, nonce, from[:], bodyBs)
	tx.Signature, err = tools.SignSecp256k1(tx, crypto.FromECDSA(key))
	if err != nil {
		return cli.NewExitError(err.Error(), 123)
	}
	b, err := tools.ToBytes(tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 123)
	}

	res := new(agtypes.ResultBroadcastTx)
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	_, err = clientJSON.Call("broadcast_tx_sync", []interface{}{agtypes.WrapTx(types.TxTagAppEvmCommon, b)}, res)
	if err != nil {
		return cli.NewExitError(err.Error(), 123)
	}

	if res.Code != 0 {
		fmt.Println("Error:", res.Log)
	} else {
		fmt.Printf("txHash: %x", tools.Hash(tx))
	}

	return nil
}

func createContract(ctx *cli.Context) error {
	json, err := getCallParamsJSON(ctx)
	if err != nil {
		cli.NewExitError(err.Error(), 127)
	}
	aabbii, err := getAbiJSON(ctx)
	if err != nil {
		cli.NewExitError(err.Error(), 127)
	}
	params := json.Get("params").MustArray()
	privkey := json.Get("privkey").MustString()
	if privkey == "" {
		privkey = json.Get("privkey").MustString()
	}
	if privkey == "" {
		return cli.NewExitError("privkey is required", 123)
	}
	key, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return cli.NewExitError(err.Error(), 110)
	}
	bytecode := common.Hex2Bytes(json.Get("bytecode").MustString())
	if len(bytecode) == 0 {
		cli.NewExitError("please give me the bytecode the contract", 127)
	}
	if len(params) > 0 {
		args, err := commons.ParseArgs("", *aabbii, params)
		if err != nil {
			cli.NewExitError(err.Error(), 127)
		}
		data, err := aabbii.Pack("", args...)
		if err != nil {
			cli.NewExitError(err.Error(), 127)
		}
		bytecode = append(bytecode, data...)
	}

	from := crypto.PubkeyToAddress(key.PublicKey)
	nonce := ctx.Uint64("nonce")

	bodyTx := &types.TxEvmCommon{
		Load: bytecode,
	}
	bodyBs, err := tools.ToBytes(bodyTx)
	if err != nil {
		return cli.NewExitError(err.Error(), 123)
	}

	tx := types.NewBlockTx(gasLimit, gasPrice, nonce, from[:], bodyBs)
	tx.Signature, err = tools.SignSecp256k1(tx, crypto.FromECDSA(key))
	if err != nil {
		return cli.NewExitError(err.Error(), 123)
	}
	b, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 123)
	}

	bytesLoad := agtypes.WrapTx(types.TxTagAppEvmCommon, b)
	res := new(agtypes.ResultBroadcastTxCommit)
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	_, err = clientJSON.Call("broadcast_tx_commit", []interface{}{bytesLoad}, res)
	if err != nil {
		return cli.NewExitError(err.Error(), 110)
	}

	if res.Code != 0 {
		fmt.Println("Error:", res.Log)
	} else {
		fmt.Printf("txHash: %x\n", tools.Hash(tx))
	}

	contractAddr := crypto.CreateAddress(from, nonce)
	fmt.Println("contract address:", contractAddr.Hex())

	return nil
}

func existContract(ctx *cli.Context) error {
	json, err := getCallParamsJSON(ctx)
	if err != nil {
		cli.NewExitError(err.Error(), 127)
	}
	bytecode := json.Get("bytecode").MustString()
	contractAddress := json.Get("contract").MustString()
	if contractAddress == "" || bytecode == "" {
		return cli.NewExitError("missing params", 127)
	}
	if strings.Contains(bytecode, "f300") {
		bytecode = strings.Split(bytecode, "f300")[1]
	}
	data := common.Hex2Bytes(bytecode)
	contractAddr := common.HexToAddress(ac.SanitizeHex(contractAddress))
	tx := ethtypes.NewTransaction(0, common.Address{}, contractAddr, big.NewInt(0), gasLimit, big.NewInt(0), crypto.Keccak256(data))

	txBytes, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	query := append([]byte{types.QueryTypeContractExistance}, txBytes...)
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	res := new(agtypes.ResultQuery)
	_, err = clientJSON.Call("query", []interface{}{query}, res)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	hex := common.Bytes2Hex(res.Result.Data)
	if hex == "01" {
		fmt.Println("Yes!!!")
	} else {
		fmt.Println("No!!!")
	}
	return nil
}

func getCallParamsJSON(ctx *cli.Context) (*simplejson.Json, error) {
	var calljson string
	if ctx.String("callf") != "" {
		dat, err := ioutil.ReadFile(ctx.String("callf"))
		if err != nil {
			return nil, err
		}
		calljson = string(dat)
	} else {
		calljson = ctx.String("calljson")
	}
	return simplejson.NewJson([]byte(calljson))
}

func getAbiJSON(ctx *cli.Context) (*abi.ABI, error) {
	var abijson string
	if ctx.String("abif") == "" {
		abijson = ctx.String("abi")
	} else {
		dat, err := ioutil.ReadFile(ctx.String("abif"))
		if err != nil {
			return nil, err
		}
		abijson = string(dat)
	}
	jAbi, err := abi.JSON(strings.NewReader(abijson))
	if err != nil {
		return nil, err
	}
	return &jAbi, nil
}

func queryContract(ctx *cli.Context) error {
	json, err := getCallParamsJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	aabbii, err := getAbiJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	function := json.Get("function").MustString()
	if !aabbii.Methods[function].Const {
		fmt.Printf("we can only read constant method, %s is not! Any consequence is on you.\n", function)
	}
	params := json.Get("params").MustArray()
	contractAddress := ac.SanitizeHex(json.Get("contract").MustString())
	args, err := commons.ParseArgs(function, *aabbii, params)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	data, err := aabbii.Pack(function, args...)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	to := common.HexToAddress(contractAddress)
	bodyTx := &types.TxEvmCommon{
		To:   to[:],
		Load: data,
	}
	bodyBs, err := tools.ToBytes(bodyTx)
	if err != nil {
		return cli.NewExitError(err.Error(), 123)
	}

	from := common.Address{}

	tx := types.NewBlockTx(gasLimit, gasPrice, 0, from[:], bodyBs)

	b, err := tools.ToBytes(tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 123)
	}
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	res := new(agtypes.ResultQuery)
	_, err = clientJSON.Call("query_contract", []interface{}{b}, res)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	// hex := common.Bytes2Hex(res.Result.Data)
	// fmt.Println("query result:", hex)
	parseResult, _ := unpackResult(function, *aabbii, string(res.Result.Data))
	fmt.Println("parse result:", parseResult)

	return nil
}
