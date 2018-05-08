package types

import (
	"math/big"

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
	Payload   []byte
	Signature []byte
}

type TxEvmCommon struct {
	To     []byte
	Amount *big.Int
	Load   []byte
}

type TxShareEco struct {
	Source    []byte
	Amount    *big.Int
	Signature []byte
}

type TxShareTransfer struct {
	ShareSrc []byte
	ShareDst []byte
	Amount   *big.Int
	ShareSig []byte
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

func (tx *BlockTx) SigObject() interface{} {
	return []interface{}{
		tx.GasLimit,
		tx.GasPrice,
		tx.Nonce,
		tx.Sender,
		tx.Payload,
	}
}

func (tx *TxShareEco) SigObject() interface{} {
	return []interface{}{
		tx.Source,
		tx.Amount,
	}
}

func (tx *TxShareTransfer) SigObject() interface{} {
	return []interface{}{
		tx.ShareSrc,
		tx.ShareDst,
		tx.Amount,
	}
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
			gas := big.Int{}
			totalGas.Add(totalGas, gas.Mul(tx.GasPrice, tx.GasLimit))
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
