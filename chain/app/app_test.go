package app

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/Baptist-Publication/chorus/module/lib/go-crypto"
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
