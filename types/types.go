package types

import (
	"math/big"
)

var (
	big0 = big.NewInt(0)

	TxTagApp          = []byte{1}
	TxTagAppEvm       = []byte{1, 1}
	TxTagAppEvmCommon = []byte{1, 1, 1}

	TxTagAppEco              = []byte{1, 2}
	TxTagAppEcoShareTransfer = []byte{1, 2, 1}
	TxTagAppEcoGuarantee     = []byte{1, 2, 2}
	TxTagAppEcoRedeem        = []byte{1, 2, 3}
)

var (
	TxTagAngine          = []byte{3} //suspect have used 0x02
	TxTagAngineInit      = []byte{3, 1}
	TxTagAngineInitToken = []byte{3, 1, 1}
	TxTagAngineInitShare = []byte{3, 1, 2}
)

const (
	CODE_VAR_ENT = "ent_params"
	CODE_VAR_RET = "ret_params"
)

func BigInt0() *big.Int {
	return big0
}

const (
	QueryTypeContract = iota
	QueryTypeNonce
	QueryTypeBalance
	QueryTypeReceipt
	QueryTypeContractExistance
	QueryTypeShare
)

type QueryShareResult struct {
	ShareBalance  *big.Int
	ShareGuaranty *big.Int
	MHeight       uint64
}
