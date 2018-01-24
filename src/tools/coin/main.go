package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"time"

	"github.com/Baptist-Publication/chorus-module/lib/eth/common"
	ethtypes "github.com/Baptist-Publication/chorus-module/lib/eth/core/types"
	ethcrypto "github.com/Baptist-Publication/chorus-module/lib/eth/crypto"
	"github.com/Baptist-Publication/chorus-module/lib/eth/rlp"

	"github.com/Baptist-Publication/angine/types"
	cl "github.com/Baptist-Publication/chorus-module/lib/go-rpc/client"
	"github.com/Baptist-Publication/chorus-module/xlog"
	"github.com/Baptist-Publication/chorus/src/tools"
)

func init() {
	xlog.Init("logs", 8)
	xlog.Info("start...")

	//config xlog
	runtime.GOMAXPROCS(runtime.NumCPU())

	xlog.Infoln("program run with %d cpus.", runtime.NumCPU())
}

type Config struct {
	Duration string `json:"duration"`
	Host     string `json:"host"`
	Value    int64  `json:"value"`
	Payload  string `json:"payload"`
	Number   int64  `json:"number"`
	ChainID  string `json:"chainid"`
}

func loadConfig(conf *Config) {
	configfile := "config.json"
	file, err := os.Open(configfile)
	if err != nil {
		xlog.Errorln("cannot find config file:", err)
		return
	}
	defer file.Close()

	jsonParse := json.NewDecoder(file)
	if err = jsonParse.Decode(&conf); err != nil {
		xlog.Errorln("cannot decode config file: ", err)
	}
}

func main() {
	flag.Parse()
	//recover
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("critical error, recover:", r)
		}
	}()
	defer xlog.Flush()

	var conf = Config{}
	loadConfig(&conf)

	duration, _ := time.ParseDuration(conf.Duration)
	fmt.Println(duration)
	concurrency := conf.Number
	ticker := time.NewTicker(duration)
	totalBatch := (tools.NumAccounts / 2) / concurrency
	batch := int64(0)
	nonce := uint64(0)
	addrCursor := int64(0)

	fmt.Println("prepare private keys...")
	privkeys := tools.PreparePrivateKeys(0, tools.NumAccounts)
	fmt.Println(len(privkeys), " private keys are ready")

	client := cl.NewClientJSONRPC(conf.Host)
	for {
		select {
		case <-ticker.C:
			for i := batch * concurrency; i < (batch+1)*concurrency; i++ {
				if int64(len(privkeys)) <= addrCursor {
					break
				}
				go wire(client, privkeys[addrCursor], ethcrypto.PubkeyToAddress(privkeys[addrCursor+1].PublicKey), nonce, conf)
				addrCursor += 2
			}
			batch++
			if batch > totalBatch {
				batch = 0
				addrCursor = 0
				nonce++
			}
		}
	}
}

func wire(client *cl.ClientJSONRPC, privkey *ecdsa.PrivateKey, toaddr common.Address, nonce uint64, conf Config) error {
	data := common.Hex2Bytes(conf.Payload)
	tx := ethtypes.NewTransaction(nonce, toaddr, big.NewInt(conf.Value), big.NewInt(90000), big.NewInt(0), data)
	sig, err := ethcrypto.Sign(tx.SigHash(tools.EthSigner).Bytes(), privkey)
	if err != nil {
		return fmt.Errorf("sign error:%v", err)
	}
	sigTx, err := tx.WithSignature(tools.EthSigner, sig)
	if err != nil {
		return fmt.Errorf("withsignature error:%v", err)
	}

	b, err := rlp.EncodeToBytes(sigTx)
	if err != nil {
		return fmt.Errorf("encode to bytes error:%v", err)
	}

	tmResult := new(types.RPCResult)
	_, err = client.Call("broadcast_tx_sync", []interface{}{conf.ChainID, b}, tmResult)
	if err != nil {
		return fmt.Errorf("send tx error:%v", err)
	}
	res := (*tmResult).(*types.ResultBroadcastTx)
	// xlog.Infoln("result data :", common.Bytes2Hex(res.Data))
	xlog.Dbg("result data :", string(res.Data))
	return nil
}
