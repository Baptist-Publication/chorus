package app

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/Baptist-Publication/chorus/eth/rlp"
	"github.com/Baptist-Publication/chorus/module/lib/go-crypto"
	db "github.com/Baptist-Publication/chorus/module/lib/go-db"
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

func TestBar(t *testing.T) {
	type xx struct {
		Name string
		Age  *big.Int
	}

	x := xx{
		Name: "lilei",
		Age:  big.NewInt(18),
	}

	b, _ := rlp.EncodeToBytes(x)
	fmt.Printf("%x\n", b)
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
	ss := getShareState()
	app.ShareState = ss

	for i := 0; i < 50; i++ {
		k := randomPubkey()
		p := new(big.Int).SetUint64(uint64(10 + i*10))
		ss.CreateShareAccount(k[:], p)
		ss.AddGuaranty(k, p, 20)
		// fmt.Printf("%X-%s\n", k[:2], p.String())
	}

	root, _ := ss.Commit()
	ss.Reload(root)

	bigbang := new(big.Int).SetBytes(root)

	for i := 0; i < 10; i++ {
		bigbang.Add(bigbang, new(big.Int).SetUint64(rand.Uint64()))
		vals := app.doElect(bigbang, 20, 2)
		// fmt.Println(len(vals))
		for _, v := range vals {
			// fmt.Printf("%s-%d ", v.PubKey.KeyString()[:4], v.VotingPower)
			fmt.Printf("%d ", v.VotingPower)
		}
		fmt.Println()
	}
}
