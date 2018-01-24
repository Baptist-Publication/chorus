package main

import (
	"fmt"

	ethcmn "github.com/Baptist-Publication/chorus-module/lib/eth/common"
	exsdk "github.com/Baptist-Publication/chorus/src/example/sdk"
	extypes "github.com/Baptist-Publication/chorus/src/example/types"
	"github.com/Baptist-Publication/chorus/src/tools"
)

const (
	PORT = ":46666"
)

type ExternalApp struct {
}

func (app *ExternalApp) CheckTx(acc ethcmn.Address, txBs []byte) error {
	var tx extypes.ExternalTx
	exsdk.ToMsg(txBs, &tx)
	fmt.Println("invoke checkTx,acc:", acc.Hex(), ",txdata:", tx.HiData)
	return nil
}

func (app *ExternalApp) Execute(helper *exsdk.Helper, blockHeight int64, acc ethcmn.Address, txBytes []byte) ([]byte, error) {
	var txData extypes.ExternalTx
	exsdk.ToMsg(txBytes, &txData)
	receipt, err := app.ServeTx(helper, acc, blockHeight, &txData)

	receipt.TxHash, err = tools.HashKeccak(txBytes)
	if err != nil {
		return nil, err
	}

	fmt.Println("invoke execute,acc:", acc.Hex(), ",height:", blockHeight, ",err:", err)

	rbytes := exsdk.ToBytes(&receipt)
	return rbytes, err
}

func (app *ExternalApp) ServeTx(h *exsdk.Helper, accid ethcmn.Address, blockHeight int64, tx *extypes.ExternalTx) (receipt extypes.ExternalReceipt, err error) {

	receipt.BlockHeight = blockHeight
	receipt.AccountID = accid.Bytes()
	receipt.Log = tx.HiData

	switch tx.HiData {
	case "add":
		err = h.AddData("hello", []byte("world"))
		fmt.Println("add,ok:", err)
	case "query":
		var rev []byte
		rev, err = h.QueryData("hello")
		fmt.Println("query rev:", string(rev), ",ok", err)
	}
	return
}

func main() {
	svr := exsdk.NewServer(&ExternalApp{})
	fmt.Println("serving...:", PORT)
	err := svr.Start(PORT)
	fmt.Println("stoped,err:", err)
}
