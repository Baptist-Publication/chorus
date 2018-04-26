package xlib

import (
	"bytes"
	"testing"

	"github.com/Baptist-Publication/chorus/module/lib/ed25519"
	crypto "github.com/Baptist-Publication/chorus/module/lib/go-crypto"
)

func TestSortInt64Slc(t *testing.T) {
	var slc = []int64{1, 129, 20, 4, 45, 66}
	Int64Slice(slc).Sort()
	pre := slc[0]
	for i := range slc {
		if pre > slc[i] {
			t.Error("sort err")
			return
		}
		pre = slc[i]
	}
}

func TestSerialBytes(t *testing.T) {
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], crypto.CRandBytes(32))
	pubKeyBytes := ed25519.MakePublicKey(privKeyBytes)
	set := (*pubKeyBytes)[:]
	var wbuffer bytes.Buffer
	if err := WriteBytes(&wbuffer, set); err != nil {
		t.Error("writeBytes err:", err)
		return
	}
	wbys := wbuffer.Bytes()

	rbuffer := bytes.NewReader(wbys)
	ret, err := ReadBytes(rbuffer)
	if err != nil {
		t.Error("readBytes err:", err)
		return
	}
	//fmt.Printf("origin:%X\nwrite bytes:%X\nre-read bytes:%X\n", set, wbys, ret)
	if string(ret) != string(set) {
		t.Error("read error, not equal")
	}
}
