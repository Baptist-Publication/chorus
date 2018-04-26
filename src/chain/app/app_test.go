package app

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
)

func TestFoo(t *testing.T) {
	s := "81e5c24b7f2de45e9a02f47110fce0c617e64ad1f3a05fed8d6acb07ba90f435"
	// p := "d011759a1169e3c47ec5f21179145b26c180205aafdeb0f83de5984861130908"
	skbs, _ := hex.DecodeString(s)

	var sk crypto.PrivKeyEd25519
	copy(sk[:], skbs)

	pk := sk.PubKey()
	fmt.Println(pk.KeyString())

	// sig := sk.Sign([]byte{1, 2, 3})

	// crypto.GenPrivKeyEd25519()

	// var pk crypto.PubKeyEd25519
	// pkbs, _ := hex.DecodeString(p)
	// copy(pk[:], pkbs)

	// if !pk.VerifyBytes([]byte{1, 2, 3}, sig) {
	// 	panic("123")
	// }
}

type base struct {
	Base int
}

func (b *base) Bar(f interface{}) string {
	s, _ := json.Marshal(b)
	return string(s)
}

type Foo struct {
	base
	Foo int
}

func TestBar(t *testing.T) {
	var f Foo
	fmt.Println(f.Bar(f))
}
