package tools

import (
	"golang.org/x/crypto/ripemd160"
)

func HashRipemd160(data []byte) ([]byte, error) {
	ripe160 := ripemd160.New()
	if _, err := ripe160.Write(data); err != nil {
		return nil, err
	}
	return ripe160.Sum(nil), nil
}
