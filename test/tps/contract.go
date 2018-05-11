package main

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	"github.com/Baptist-Publication/chorus/eth/accounts/abi"
	"github.com/Baptist-Publication/chorus/eth/common"
	ethtypes "github.com/Baptist-Publication/chorus/eth/core/types"
	"github.com/Baptist-Publication/chorus/eth/crypto"
	"github.com/Baptist-Publication/chorus/eth/rlp"
	ac "github.com/Baptist-Publication/chorus/module/lib/go-common"
	cl "github.com/Baptist-Publication/chorus/module/lib/go-rpc/client"
	"github.com/Baptist-Publication/chorus/tools"
	"github.com/Baptist-Publication/chorus/types"
)

func createContract(client *cl.ClientJSONRPC, privkey, bytecode string, nonce uint64) (string, error) {
	sk, err := crypto.HexToECDSA(ac.SanitizeHex(privkey))
	panicErr(err)

	btxbs, err := tools.ToBytes(&types.TxEvmCommon{
		Amount: new(big.Int),
		Load:   common.Hex2Bytes(bytecode),
	})
	panicErr(err)

	tx := types.NewBlockTx(gasLimit, big.NewInt(0), nonce, crypto.PubkeyToAddress(sk.PublicKey).Bytes(), btxbs)
	tx.Signature, err = tools.SignSecp256k1(tx, crypto.FromECDSA(sk))
	panicErr(err)
	b, err := tools.ToBytes(tx)
	panicErr(err)

	// tmResult := new(agtypes.TMResult)
	tmResult := new(agtypes.RPCResult)
	if client == nil {
		client = cl.NewClientJSONRPC(logger, rpcTarget)
	}
	_, err = client.Call("broadcast_tx_sync", []interface{}{append(types.TxTagAppEvmCommon, b...)}, tmResult)
	panicErr(err)

	res := (*tmResult).(*agtypes.ResultBroadcastTx)
	if res.Code != 0 {
		fmt.Println(res.Code, string(res.Data))
		return "", errors.New(string(res.Data))
	}

	fmt.Println(res.Code, string(res.Data))
	caller := crypto.PubkeyToAddress(crypto.ToECDSA(common.Hex2Bytes(privkey)).PublicKey)
	addr := crypto.CreateAddress(caller, nonce)
	fmt.Println("contract address:", addr.Hex())

	return addr.Hex(), nil
}

func executeContract(client *cl.ClientJSONRPC, privkey, contract, abijson, callfunc string, args []interface{}, nonce uint64) error {
	sk, err := crypto.HexToECDSA(ac.SanitizeHex(privkey))
	panicErr(err)

	aabbii, err := abi.JSON(strings.NewReader(abijson))
	panicErr(err)
	calldata, err := aabbii.Pack(callfunc, args...)
	panicErr(err)

	btxbs, err := tools.ToBytes(&types.TxEvmCommon{
		To:     common.HexToAddress(contract).Bytes(),
		Amount: new(big.Int),
		Load:   calldata,
	})
	panicErr(err)

	tx := types.NewBlockTx(gasLimit, big.NewInt(0), nonce, crypto.PubkeyToAddress(sk.PublicKey).Bytes(), btxbs)
	tx.Signature, err = tools.SignSecp256k1(tx, crypto.FromECDSA(sk))
	panicErr(err)
	b, err := tools.ToBytes(tx)
	panicErr(err)

	tmResult := new(agtypes.RPCResult)
	if client == nil {
		client = cl.NewClientJSONRPC(logger, rpcTarget)
	}
	_, err = client.Call("broadcast_tx_sync", []interface{}{append(types.TxTagAppEvmCommon, b...)}, tmResult)
	panicErr(err)

	res := (*tmResult).(*agtypes.ResultBroadcastTx)
	if res.Code != 0 {
		fmt.Println(res.Code, string(res.Data), res.Log)
		return errors.New(string(res.Data))
	}

	return nil
}

func readContract(client *cl.ClientJSONRPC, privkey, contract, abijson, callfunc string, args []interface{}, nonce uint64) error {
	sk, err := crypto.HexToECDSA(privkey)
	panicErr(err)

	aabbii, err := abi.JSON(strings.NewReader(abijson))
	panicErr(err)
	data, err := aabbii.Pack(callfunc, args...)
	panicErr(err)

	// nonce := uint64(time.Now().UnixNano())
	to := common.HexToAddress(contract)
	privkey = ac.SanitizeHex(privkey)

	tx := ethtypes.NewTransaction(nonce, crypto.PubkeyToAddress(sk.PublicKey), to, big.NewInt(0), gasLimit, big.NewInt(0), data)
	b, err := rlp.EncodeToBytes(tx)
	panicErr(err)

	query := append([]byte{types.QueryTypeContract}, b...)
	tmResult := new(agtypes.RPCResult)
	if client == nil {
		client = cl.NewClientJSONRPC(logger, rpcTarget)
	}
	_, err = client.Call("query", []interface{}{query}, tmResult)
	panicErr(err)

	res := (*tmResult).(*agtypes.ResultQuery)
	fmt.Println("query result:", common.Bytes2Hex(res.Result.Data))
	parseResult, _ := unpackResult(callfunc, aabbii, string(res.Result.Data))
	fmt.Println("parse result:", reflect.TypeOf(parseResult), parseResult)

	return nil
}

func existContract(client *cl.ClientJSONRPC, privkey, contract, bytecode string) bool {
	sk, err := crypto.HexToECDSA(privkey)
	panicErr(err)

	if strings.Contains(bytecode, "f300") {
		bytecode = strings.Split(bytecode, "f300")[1]
	}

	data := common.Hex2Bytes(bytecode)
	privkey = ac.SanitizeHex(privkey)
	to := common.HexToAddress(ac.SanitizeHex(contract))

	tx := ethtypes.NewTransaction(0, crypto.PubkeyToAddress(sk.PublicKey), to, big.NewInt(0), gasLimit, big.NewInt(0), crypto.Keccak256(data))
	b, err := rlp.EncodeToBytes(tx)
	panicErr(err)

	query := append([]byte{types.QueryTypeContractExistance}, b...)
	tmResult := new(agtypes.RPCResult)
	if client == nil {
		client = cl.NewClientJSONRPC(logger, rpcTarget)
	}
	_, err = client.Call("query", []interface{}{query}, tmResult)
	panicErr(err)

	res := (*tmResult).(*agtypes.ResultQuery)
	hex := common.Bytes2Hex(res.Result.Data)
	if hex == "01" {
		return true
	}
	return false
}
