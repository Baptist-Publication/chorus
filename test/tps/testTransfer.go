package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Baptist-Publication/chorus/eth/common"
	"github.com/Baptist-Publication/chorus/eth/crypto"
	cl "github.com/Baptist-Publication/chorus/module/lib/go-rpc/client"
)

var (
	txAmount int64 = 0
)

func testTxCallOnce() {
	client := cl.NewClientJSONRPC(logger, rpcTarget)

	pk := crypto.ToECDSA(common.Hex2Bytes(defaultPrivKey))
	caller := crypto.PubkeyToAddress(pk.PublicKey)

	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)

	err = send(client, defaultPrivKey, defaultReceiver, txAmount, nonce)
	panicErr(err)
}

func testPushTx() {
	fmt.Println("ThreadCount:", threadCount)
	fmt.Println("SendPerThread:", sendPerThread)
	time.Sleep(time.Second * 2)

	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup

	go resPrintRoutine()

	for i := 0; i < threadCount-1; i++ {
		go testTx(&wg, i, fmt.Sprintf("%06dCD0D48031A21F4B50EBDE558CE5294C550390118C87A3E8C69DCAFE89A", rand.Uint64()%1000000))
	}

	testTx(&wg, threadCount-1, fmt.Sprintf("%06dCD0D48031A21F4B50EBDE558CE5294C550390118C87A3E8C69DCAFE89A", rand.Uint64()%1000000)) // use to block routine

	wg.Wait()
}

func testTx(w *sync.WaitGroup, id int, privkey string) {
	if w != nil {
		w.Add(1)
	}

	fmt.Println("using privkey:", privkey)
	time.Sleep(time.Second * 2)

	if privkey == "" {
		privkey = defaultPrivKey
	}
	client := cl.NewClientJSONRPC(logger, rpcTarget)

	pk := crypto.ToECDSA(common.Hex2Bytes(privkey))
	caller := crypto.PubkeyToAddress(pk.PublicKey)

	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)

	for i := 0; i < sendPerThread; i++ {
		err := send(client, privkey, defaultReceiver, 0, nonce)
		panicErr(err)

		resq <- res{id, sendPerThread - i}
		time.Sleep(time.Millisecond * time.Duration(sleep))

		nonce++
	}

	if w != nil {
		w.Done()
	}
}

func showReceiverBalance() {
	balance, err := getBalance(nil, defaultReceiver)
	panicErr(err)

	fmt.Println("balance:", balance)
}
