package app

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"testing"

	"github.com/Baptist-Publication/chorus/config"
	"github.com/Baptist-Publication/chorus/module/lib/go-crypto"
	db "github.com/Baptist-Publication/chorus/module/lib/go-db"
	"github.com/Baptist-Publication/chorus/tools"
	"github.com/Baptist-Publication/chorus/types"
)

func TestFoo(t *testing.T) {
	content := []byte{1, 2, 3}
	skstr := "ED30BE21F0C6F97C4656EA17680C51F9867B30D40FEEDC81C70C54DF11C9C435A18EB771ACFCA6706A40C7FD419D0081011811403EAA0B75DB5D9B3D1A73F288"
	pkstr := "A18EB771ACFCA6706A40C7FD419D0081011811403EAA0B75DB5D9B3D1A73F288"
	skbs, _ := hex.DecodeString(skstr)
	pkbs, _ := hex.DecodeString(pkstr)

	var sk crypto.PrivKeyEd25519
	var pk crypto.PubKeyEd25519
	copy(sk[:], skbs)
	copy(pk[:], pkbs)

	fmt.Printf("Calculated pubkey: %x\n", sk.PubKey().Bytes())
	fmt.Printf("%x\n", sk.Bytes())

	// pk := sk.PubKey()
	// fmt.Println(pk.KeyString())

	sig := sk.Sign(content)

	if !pk.VerifyBytes(content, sig) {
		t.Fail()
	}
}

func toBytes(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}

func TestBar(t *testing.T) {
	evmTx := &types.TxEvmCommon{
		To:     toBytes("00000000000000000000000000000000000000c1"),
		Amount: big.NewInt(1000),
		// Load:   []byte(""),
	}
	evmBs, err := tools.ToBytes(evmTx)
	if err != nil {
		panic(err)
	}

	tx := types.NewBlockTx(big.NewInt(1000000), big.NewInt(1), 1, toBytes("7a96091f2abb51ba304ad6be6043e696ada5210b"), evmBs)
	pkbs, _ := hex.DecodeString("760b490afd1341f1aaacd8b1bcc97d5cc84d21f085c3d4935c38816cee637ab7")

	tx.Signature, err = tools.SignSecp256k1(tx, pkbs)
	if err != nil {
		panic(err)
	}

	if err = tools.VerifySecp256k1(tx, tx.Sender, tx.Signature); err != nil {
		panic(err)
	}

	// bs, err := tools.ToBytes(tx)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Printf("%x\n", bs)
}

func ToJSON(o interface{}) string {
	bs, _ := json.Marshal(o)
	return string(bs)
}

func TestNothing(t *testing.T) {
	s := "f877830186a001809471ac67bef29722c49b01b7b23c88663a1bbe575c98d79400a3553efce903931811ebc25871a626cf89e7086480b841459320b3897863131af6fe8ec9343f89abf751fe64d3a4070f256ad50848ce961a0638a19c3a42de9725b182d6e910aaaabe3e2bfc7fb3f6605cfcca86dc913e00"
	bs, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}

	tx := &types.BlockTx{}
	if err = tools.FromBytes(bs, tx); err != nil {
		panic(err)
	}

	fmt.Println(ToJSON(tx))

	fmt.Println(len(tx.Signature))

	fmt.Printf("%x\n", tx.Signature)
	tx.Signature[64] = 1

	if err = tools.VerifySecp256k1(tx, tx.Sender, tx.Signature); err != nil {
		panic(err)
	}

	bodyTx := &types.TxShareTransfer{}
	if err = tools.FromBytes(tx.Payload, bodyTx); err != nil {
		panic(err)
	}

	fmt.Println(ToJSON(bodyTx))

	if err = tools.VeirfyED25519(bodyTx, bodyTx.ShareSrc, bodyTx.ShareSig); err != nil {
		panic(err)
	}

}

func randomPubkey() *crypto.PubKeyEd25519 {
	sk1 := crypto.GenPrivKeyEd25519()
	pk2 := sk1.PubKey()
	pk2bs := pk2.(*crypto.PubKeyEd25519)

	return pk2bs
}

func randomBig(max uint64) *big.Int {
	return new(big.Int).SetUint64(rand.Uint64() % max)
}

func getShareState() *ShareState {
	shrDB, err := db.NewGoLevelDB(fmt.Sprintf("sharestate%d", rand.Int()), "/tmp/")
	if err != nil {
		panic(err)
	}

	return NewShareState(shrDB)
}

func TestElection(t *testing.T) {
	app := &App{}
	app.Config = config.GetConfig("")
	app.Config.Set("elect_threshold_lucky", 10000)
	app.Config.Set("elect_threshold_rich", 1000000)
	ss := getShareState()
	app.ShareState = ss

	// for i := 0; i < 50; i++ {
	// 	k := randomPubkey()
	// 	p := new(big.Int).SetUint64(uint64(1000000 + i*10))
	// 	ss.AddGuaranty(k, p, 20)
	// }

	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(2000000), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(20013), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(20555), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(20111), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(20335), 20)
	ss.AddGuaranty(randomPubkey(), big.NewInt(20888), 20)

	root, _ := ss.Commit()
	ss.Reload(root)

	fmt.Println(ss.Size())

	bigbang := new(big.Int).SetBytes(root)

	for i := 0; i < 1; i++ {
		bigbang.Add(bigbang, new(big.Int).SetUint64(rand.Uint64()))
		vals := app.doElect(bigbang, 20, 1)
		// fmt.Println(len(vals))
		for _, v := range vals {
			fmt.Printf("%s-%d ", v.PubKey.KeyString()[:4], v.VotingPower)
			// fmt.Printf("%d ", v.VotingPower)
		}
		fmt.Println()
	}
}

func TestJSON(t *testing.T) {
	type foo struct {
		Name string
		Age  int
	}

	f := foo{
		Name: "lilei",
		Age:  18,
	}

	b, err := json.Marshal(f)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	s := "{'Name':'lilei','Age':18}"
	s = strings.Replace(s, "'", "\"", -1)
	f1 := &foo{}
	err = json.Unmarshal([]byte(s), f1)
	if err != nil {
		panic(err)
	}

}
