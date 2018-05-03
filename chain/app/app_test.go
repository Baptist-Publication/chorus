package app

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/Baptist-Publication/chorus/module/lib/go-crypto"
	db "github.com/Baptist-Publication/chorus/module/lib/go-db"
)

func TestFoo(t *testing.T) {
	s := "81e5c24b7f2de45e9a02f47110fce0c617e64ad1f3a05fed8d6acb07ba90f435"
	p := "d011759a1169e3c47ec5f21179145b26c180205aafdeb0f83de5984861130908"
	skbs, _ := hex.DecodeString(s)

	var sk crypto.PrivKeyEd25519
	copy(sk[:], skbs)

	// pk := sk.PubKey()
	// fmt.Println(pk.KeyString())

	sig := sk.Sign([]byte{1, 2, 3})

	fmt.Println(sig)

	crypto.GenPrivKeyEd25519()

	var pk crypto.PubKeyEd25519
	pkbs, _ := hex.DecodeString(p)
	copy(pk[:], pkbs)

	if !pk.VerifyBytes([]byte{1, 2, 3}, sig) {
		panic("123")
	}
}

func TestBar(t *testing.T) {
	// base := make([]byte, 3, 4)
	// base[0] = 1
	// base[1] = 2
	// base[2] = 3
	base := []byte{1, 2, 3}

	d1 := append(base, 4)
	d2 := append(base, 5)

	fmt.Println(base)
	fmt.Println(d1)
	fmt.Println(d2)
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
