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

type res struct {
	id   int
	left int
}

var (
	threadCount   = 10
	sendPerThread = 1000
)

var (
	resq   = make(chan res, 1024)
	resarr = make([]int, threadCount)

	ops int
)

func testContractCallOnce() {
	assertContractExist(nil)

	client := cl.NewClientJSONRPC(logger, rpcTarget)

	callFunc := "add"
	args := []interface{}{}
	pk := crypto.ToECDSA(common.Hex2Bytes(defaultPrivKey))
	caller := crypto.PubkeyToAddress(pk.PublicKey)

	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)

	err = executeContract(client, defaultPrivKey, defaultContractAddr, defaultAbis, callFunc, args, nonce)
	panicErr(err)
}

func testPushContract() {
	assertContractExist(nil)

	fmt.Println("ThreadCount:", threadCount)
	fmt.Println("SendPerThread:", sendPerThread)
	time.Sleep(time.Second * 2)

	var wg sync.WaitGroup

	go resPrintRoutine()

	for i := 0; i < threadCount-1; i++ {
		go testContract(&wg, i, fmt.Sprintf("7d73c3dafd3c0215b8526b26f8dbdb93242fc7dcfbdfa1000d93436d577c%04d", rand.Uint64()%10000))
		// go testContract(&wg,i, "")
	}

	testContract(&wg, threadCount-1, "") // use to block routine

	wg.Wait()
}

func testContract(w *sync.WaitGroup, id int, privkey string) {
	// var err error
	if w != nil {
		w.Add(1)
	}

	if privkey == "" {
		privkey = defaultPrivKey
	}
	client := cl.NewClientJSONRPC(logger, rpcTarget)

	callFunc := "add"
	args := []interface{}{}
	pk := crypto.ToECDSA(common.Hex2Bytes(privkey))
	caller := crypto.PubkeyToAddress(pk.PublicKey)

	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)

	for i := 0; i < sendPerThread; i++ {
		// nonce, err := getNonce(client, caller.Hex())
		// panicErr(err)

		// nonce := uint64(time.Now().UnixNano())
		err := executeContract(client, privkey, defaultContractAddr, defaultAbis, callFunc, args, nonce)
		panicErr(err)

		// fmt.Printf("%d: %d\n", id, sendPerThread-i)
		resq <- res{id, sendPerThread - i}
		time.Sleep(time.Millisecond * 10)

		nonce++
	}

	if w != nil {
		w.Done()
	}
}

func resPrintRoutine() {
	count := 0
	timet := time.NewTicker(time.Second)
	for {
		select {
		case r := <-resq:
			resarr[r.id] = r.left

			func() {
				s := fmt.Sprintf("[%4d op/s] ", ops)
				for i := 0; i < threadCount; i++ {
					s += fmt.Sprintf("[%02d:%4d]", i+1, resarr[i])
				}
				fmt.Println(s)
			}()

			count++
		case <-timet.C:
			ops = count
			count = 0
		}
	}
}

func testCreateContract() {
	privkey := defaultPrivKey
	client := cl.NewClientJSONRPC(logger, rpcTarget)

	pk := crypto.ToECDSA(common.Hex2Bytes(privkey))
	caller := crypto.PubkeyToAddress(pk.PublicKey)
	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)

	fmt.Println("nonce", nonce)

	caddr, err := createContract(client, privkey, defaultBytecode, nonce)
	panicErr(err)

	time.Sleep(time.Second * 3)

	exist := existContract(client, defaultPrivKey, caddr, defaultBytecode)
	fmt.Println("Contract exist:", exist)
}

func testExistContract() bool {
	client := cl.NewClientJSONRPC(logger, rpcTarget)

	exist := existContract(client, defaultPrivKey, defaultContractAddr, defaultBytecode)
	fmt.Println("Contract exist:", exist)
	return exist
}

func testReadContract() {
	var err error
	privkey := defaultPrivKey
	client := cl.NewClientJSONRPC(logger, rpcTarget)

	assertContractExist(client)

	callFunc := "get"
	args := []interface{}{}
	pk := crypto.ToECDSA(common.Hex2Bytes(privkey))
	caller := crypto.PubkeyToAddress(pk.PublicKey)
	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)

	// fmt.Println("nonce", nonce)
	err = readContract(client, privkey, defaultContractAddr, defaultAbis, callFunc, args, nonce)
	panicErr(err)
}
