package app

import (
	"fmt"
	"testing"

	"github.com/Baptist-Publication/chorus/src/eth/common"
	"github.com/Baptist-Publication/chorus/src/eth/crypto"
)

func TestFoo(t *testing.T) {
	s := "1956be6c7e5128bd1c44c389ba21bd63cfb054d5adb2ab3f8d968f74b9bd0b6b"

	sk := crypto.ToECDSA(common.Hex2Bytes(s))

	addr := crypto.PubkeyToAddress(sk.PublicKey)
	fmt.Println(addr.Hex())
}
