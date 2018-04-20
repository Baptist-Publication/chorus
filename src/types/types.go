package types

import (
	"math/big"
)

var (
	big0 = big.NewInt(0)

	TxTagApp          = []byte{1}
	TxTagAppEvm       = append(TxTagApp, 0x01)
	TxTagAppEvmCommon = append(TxTagAppEvm, 0x01)

	TxTagAppInit      = append(TxTagApp, 0x02)
	TxTagAppInitToken = append(TxTagAppInit, 0x01)
	TxTagAppInitShare = append(TxTagAppInit, 0x02)

	TxTagAppEco = append(TxTagApp, 0x03)
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
