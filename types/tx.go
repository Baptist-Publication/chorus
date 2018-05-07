package types

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"

	"github.com/Baptist-Publication/chorus/eth/crypto"
	"github.com/Baptist-Publication/chorus/eth/rlp"
	"github.com/Baptist-Publication/chorus/module/lib/ed25519"
	gcrypto "github.com/Baptist-Publication/chorus/module/lib/go-crypto"
	"sort"
)

type EcoInitTokenTx struct {
	To     []byte   `json:"to"`
	Amount *big.Int `json:"amount"`
	Extra  []byte   `json:"extra"`
}

type EcoInitShareTx struct {
	To     []byte   `json:"to"`
	Amount *big.Int `json:"amount"`
	Extra  []byte   `json:"extra"`
}

type WorldRandTx struct {
	Height uint64
	Pubkey []byte
	Sig    []byte
}

type BlockTx struct {
	GasLimit  *big.Int
	GasPrice  *big.Int
	Nonce     uint64
	Sender    []byte
	Signature []byte
	Payload   []byte
}

func NewBlockTx(gasLimit, gasPrice *big.Int, nonce uint64, sender, payload []byte) *BlockTx {
	return &BlockTx{
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Nonce:    nonce,
		Sender:   sender,
		Payload:  payload,
	}
}

type TxEvmCommon struct {
	To     []byte
	Amount *big.Int
	Load   []byte
}

type TxShareItf interface {
	Sign(*gcrypto.PrivKeyEd25519) error
	TxtoBytes() ([]byte, error)
	VerifySig() (bool, error)
}

type TxShareEco struct {
	Source    []byte
	Signature []byte
	Amount    *big.Int
}

func (tx *TxShareEco) TxtoBytes() ([]byte, error) {
	return json.Marshal(tx)
}

func (tx *TxShareEco) Sign(privkey *gcrypto.PrivKeyEd25519) error {
	txbs, err := tx.TxtoBytes()
	if err != nil {
		return err
	}
	sig := privkey.Sign(txbs).(*gcrypto.SignatureEd25519)
	tx.Signature = sig[:]
	return nil
}

func (tx *TxShareEco) VerifySig() (bool, error) {
	pubkey := gcrypto.PubKeyEd25519{}
	copy(pubkey[:], tx.Source)
	signatrue := gcrypto.SignatureEd25519{}
	copy(signatrue[:], tx.Signature)
	tx.Signature = nil
	txbs, err := tx.TxtoBytes()
	if err != nil {
		return false, err
	}
	sig64 := [64]byte(signatrue)
	pub32 := [32]byte(pubkey)
	return ed25519.Verify(&pub32, txbs, &sig64), nil
}

type TxShareTransfer struct {
	ShareSrc []byte
	ShareSig []byte
	ShareDst []byte
	Amount   *big.Int
}

func (tx *TxShareTransfer) TxtoBytes() ([]byte, error) {
	return json.Marshal(tx)
}

func (tx *TxShareTransfer) Sign(privkey *gcrypto.PrivKeyEd25519) error {
	txbs, err := tx.TxtoBytes()
	if err != nil {
		return err
	}
	sig := privkey.Sign(txbs).(*gcrypto.SignatureEd25519)
	tx.ShareSig = sig[:]
	return nil
}

func (tx *TxShareTransfer) VerifySig() (bool, error) {
	pubkey := gcrypto.PubKeyEd25519{}
	copy(pubkey[:], tx.ShareSrc)
	signatrue := gcrypto.SignatureEd25519{}
	copy(signatrue[:], tx.ShareSig)
	tx.ShareSig = nil
	txbs, err := tx.TxtoBytes()
	if err != nil {
		return false, err
	}
	sig64 := [64]byte(signatrue)
	pub32 := [32]byte(pubkey)
	return ed25519.Verify(&pub32, txbs, &sig64), nil
}

func sigHash(tx *BlockTx) ([]byte, error) {
	txbytes, err := rlp.EncodeToBytes([]interface{}{
		tx.GasLimit,
		tx.GasPrice,
		tx.Nonce,
		tx.Sender,
		tx.Payload,
	})
	if err != nil {
		return nil, err
	}

	h := crypto.Sha256(txbytes)
	return h, nil
}

func (tx *BlockTx) Sign(privkey *ecdsa.PrivateKey) error {
	h, err := sigHash(tx)
	if err != nil {
		return err
	}

	sig, err := crypto.Sign(h, privkey)
	if err != nil {
		return err
	}

	tx.Signature = sig
	return nil
}

func (tx *BlockTx) VerifySignature() (bool, error) {
	if len(tx.Signature) == 0 {
		return false, nil
	}

	h, err := sigHash(tx)
	if err != nil {
		return false, err
	}

	pub, err := crypto.Ecrecover(h, tx.Signature)
	if err != nil {
		return false, err
	}
	addr := crypto.Keccak256(pub[1:])[12:]
	return bytes.Equal(tx.Sender, addr), nil
}

func (tx *BlockTx) Hash() []byte {
	bs, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil
	}

	return crypto.Sha256(bs)
}

//sort BlockTx
type BlockTxsToSort []BlockTx

func (txs BlockTxsToSort) Len() int {
	return len(txs)
}

func (txs BlockTxsToSort) Swap(i, j int) {
	txs[i], txs[j] = txs[j], txs[i]
}

func (txs BlockTxsToSort) Less(i, j int) bool {
	return txs[i].Nonce < txs[j].Nonce
}

// collectTxs generates a 'sender:BlockTxs' pair
func collectTxs(txs []BlockTx) map[string][]BlockTx {
	if txs == nil || len(txs) == 0 {
		return nil
	}
	m := make(map[string][]BlockTx)
	for _, tx := range txs {
		m[string(tx.Sender)] = append(m[string(tx.Sender)], tx)
	}
	for sender := range m {
		sort.Sort(BlockTxsToSort(m[sender]))
	}
	return m
}

// makeSenderGases generate descending senderGas
func makeSenderGases(m map[string][]BlockTx) []senderGas {
	if m == nil || len(m) == 0 {
		return nil
	}
	var sgs []senderGas
	for sender, txs := range m {
		txLen := len(txs)
		if txLen == 0 {
			continue
		}
		totalGas := big.NewInt(0)
		for _, tx := range txs {
			totalGas.Add(totalGas, tx.GasPrice.Mul(tx.GasPrice, tx.GasLimit))
		}
		averageGas := totalGas.Div(totalGas, big.NewInt(int64(txLen)))
		sg := senderGas{sender, averageGas}
		sgs = append(sgs, sg)
	}
	sort.Sort(senderGasToSort(sgs))
	return sgs
}

// SortTxs sorts txs
func SortTxs(txs []BlockTx) []BlockTx {
	if txs == nil || len(txs) == 0 {
		return nil
	}

	m := collectTxs(txs)
	if m == nil || len(m) == 0 {
		return nil
	}

	sgs := makeSenderGases(m)
	if sgs == nil || len(sgs) == 0 {
		return nil
	}

	var ret []BlockTx
	for _, sg := range sgs {
		ret = append(ret, m[sg.sender]...)
	}
	return ret
}

type senderGas struct {
	sender string
	gas    *big.Int
}

type senderGasToSort []senderGas

func (sg senderGasToSort) Len() int {
	return len(sg)
}

func (sg senderGasToSort) Swap(i, j int) {
	sg[i], sg[j] = sg[j], sg[i]
}

func (sg senderGasToSort) Less(i, j int) bool {
	return sg[i].gas.Cmp(sg[j].gas) > 0
}
