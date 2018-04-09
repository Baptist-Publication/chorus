package main

import (
	"encoding/binary"
	"encoding/hex"

	agtypes "github.com/Baptist-Publication/angine/types"
	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	cl "github.com/Baptist-Publication/chorus-module/lib/go-rpc/client"
	"github.com/Baptist-Publication/chorus/src/chain/node"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func newClient(server string) *cl.ClientJSONRPC {
	return cl.NewClientJSONRPC(logger, server)
}

func getClient(cli *cl.ClientJSONRPC) *cl.ClientJSONRPC {
	if cli == nil {
		return newClient(defaultRPCServer)
	}
	return cli
}

func queryNonce(cli *cl.ClientJSONRPC, who []byte) uint64 {
	cli = getClient(cli)
	query := append([]byte{node.QueryNonce}, who...)
	tmResult := new(agtypes.RPCResult)

	_, err := cli.Call("query", []interface{}{defaultChainID, query}, tmResult)
	panicErr(err)

	res := (*tmResult).(*agtypes.ResultQuery)

	return binary.LittleEndian.Uint64(res.Result.Data)
}

func jsonRPC(cli *cl.ClientJSONRPC, p []byte) (*agtypes.RPCResult, error) {
	tmResult := new(agtypes.RPCResult)
	_, err := cli.Call("broadcast_tx_sync", []interface{}{defaultChainID, p}, tmResult)
	if err != nil {
		return nil, err
	}
	return tmResult, nil
}

func toPrivkey(bs string) crypto.PrivKeyEd25519 {
	skbs, err := hex.DecodeString(bs)
	panicErr(err)

	var sk crypto.PrivKeyEd25519
	copy(sk[:], skbs)

	return sk
}
