package types

import (
	"math/big"
)

var (
	big0 = big.NewInt(0)

	TxTagApp          = []byte{0x01}
	TxTagAppEvm       = append(TxTagApp, 0x01)
	TxTagAppEvmCommon = append(TxTagAppEvm, 0x01)

	TxTagAppEco              = append(TxTagApp, 0x02)
	TxTagAppEcoShareTransfer = append(TxTagAppEco, 0x01)
	TxTagAppEcoGuarantee     = append(TxTagAppEco, 0x02)
	TxTagAppEcoRedeem        = append(TxTagAppEco, 0x03)
)

var (
	TxTagAngine          = []byte{0x02}
	TxTagAngineInit      = append(TxTagAngine, 0x01)
	TxTagAngineInitToken = append(TxTagAngineInit, 0x01)
	TxTagAngineInitShare = append(TxTagAngineInit, 0x02)
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
