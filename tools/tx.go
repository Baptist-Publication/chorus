package tools

import (
	"bytes"
	"fmt"

	ethcrypto "github.com/Baptist-Publication/chorus/eth/crypto"
	"github.com/Baptist-Publication/chorus/eth/rlp"
	tmcrypto "github.com/Baptist-Publication/chorus/module/lib/go-crypto"
)

func ToBytes(tx interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(tx)
}

func FromBytes(bs []byte, tx interface{}) error {
	return rlp.DecodeBytes(bs, tx)
}

func Hash(o interface{}) []byte {
	bs, err := ToBytes(o)
	if err != nil {
		panic(err)
	}

	return ethcrypto.Sha256(bs)
}

type CanSign interface {
	SigObject() interface{}
}

func SigHash(sigObj CanSign) ([]byte, error) {
	return ToBytes(sigObj.SigObject())
}

func SignSecp256k1(tx CanSign, privkey []byte) ([]byte, error) {
	h, err := SigHash(tx)
	if err != nil {
		return nil, err
	}

	sk := ethcrypto.ToECDSA(privkey)
	sig, err := ethcrypto.Sign(h, sk)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func VerifySecp256k1(tx CanSign, sender, sig []byte) error {
	if len(sig) == 0 {
		return fmt.Errorf("empty signature")
	}

	h, err := SigHash(tx)
	if err != nil {
		return err
	}

	pub, err := ethcrypto.Ecrecover(h, sig)
	if err != nil {
		return err
	}

	if addr := ethcrypto.Keccak256(pub[1:])[12:]; !bytes.Equal(sender, addr) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func SignED25519(tx CanSign, privkey []byte) ([]byte, error) {
	h, err := SigHash(tx)
	if err != nil {
		return nil, err
	}

	sk := tmcrypto.PrivKeyEd25519{}
	copy(sk[:], privkey)

	return sk.Sign(h).Bytes(), nil
}

func VeirfyED25519(tx CanSign, sender, sig []byte) error {
	if len(sig) == 0 {
		return fmt.Errorf("empty signature")
	}

	h, err := SigHash(tx)
	if err != nil {
		return err
	}

	pk := &tmcrypto.PubKeyEd25519{}
	copy(pk[:], sender)

	s := &tmcrypto.SignatureEd25519{}
	copy(s[:], sig)

	if !pk.VerifyBytes(h, s) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}
