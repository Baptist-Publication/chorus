package main

import (
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

var (
	gasLimit = big.NewInt(1000000)

	logger *zap.Logger

	tps = 100
)

var (
	rpcTarget           = "tcp://0.0.0.0:46657"
	defaultAbis         = "[{\"constant\":false,\"inputs\":[],\"name\":\"add\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"name\":\"\",\"type\":\"int32\"}],\"payable\":false,\"type\":\"function\"}]"
	defaultBytecode     = "6060604052341561000f57600080fd5b5b6101058061001f6000396000f30060606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680634f2be91f1460475780636d4ce63c146059575b600080fd5b3415605157600080fd5b60576085565b005b3415606357600080fd5b606960c2565b604051808260030b60030b815260200191505060405180910390f35b60008081819054906101000a900460030b8092919060010191906101000a81548163ffffffff021916908360030b63ffffffff160217905550505b565b60008060009054906101000a900460030b90505b905600a165627a7a72305820259a0a3f2a8a112df2232529a36c75cc314d05060713c663a0786913fee723160029"
	defaultContractAddr = "3ffae651a8238796001e89a21d5fd15adc92e5d8"
	defaultPrivKey      = "4B4457C8E3548C970E788CB78DA2BDF739EE0E9DA2B41FA5D46E72A57469E636"
	defaultReceiver     = "7AB225A9AB3CF695DB1EA8B36D6BB9072FB798E6"
)

func main() {
	if len(os.Args) < 2 {
		panic("usage: exe type(evm、coin) op")
	}
	prepare()
	start := time.Now()

	t := os.Args[1]
	switch t {
	case "evm":
		processEVM()
	case "tx":
		processTransfer()
	case "env":
		showEnv()
	default:
		panic("unsupport type:" + t)
	}

	end := time.Now()
	fmt.Println("time used:", end.Sub(start).Seconds(), "s")
}

func processEVM() {
	showEnv()
	op := os.Args[2]
	switch op {
	case "create":
		testCreateContract()
	case "read":
		testReadContract()
	case "call":
		testContractCallOnce()
	case "push":
		time.Sleep(time.Second * 2)
		testPushContract()
	case "exist":
		testExistContract()
	default:
		panic("unsupport op:" + op)
	}
}

func processTransfer() {
	showEnv()
	op := os.Args[2]
	switch op {
	case "call":
		testTxCallOnce()
	case "push":
		time.Sleep(time.Second * 2)
		testPushTx()
	case "bal":
		showReceiverBalance()
	default:
		panic("unsupport op:" + op)
	}
}

func showEnv() {
	fmt.Println("rpc   :", rpcTarget)
	fmt.Println("thread:", threadCount)
	fmt.Println("tps   :", tps)
	// fmt.Println("per   :", sendPerThread)
	// fmt.Println("amount:", txAmount)
}

func prepare() {
	rpc := os.Getenv("rpc")
	if rpc != "" {
		rpcTarget = rpc
	}

	t := os.Getenv("thread")
	if t != "" {
		ti, _ := strconv.Atoi(t)
		threadCount = ti
	}

	per := os.Getenv("per")
	if per != "" {
		peri, _ := strconv.Atoi(per)
		sendPerThread = peri
	}

	a := os.Getenv("amount")
	if a != "" {
		ai, _ := strconv.Atoi(a)
		txAmount = int64(ai)
	}

	s := os.Getenv("tps")
	if s != "" {
		si, _ := strconv.Atoi(s)
		tps = si
	}
}
