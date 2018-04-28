package types

import (
	"math/big"
)

var (
	big0 = big.NewInt(0)
)

var (
	TxTagApp = []byte{1}

	TxTagAppEvm       = []byte{1, 1}
	TxTagAppEvmCommon = []byte{1, 1, 1}

	TxTagAppEco              = []byte{1, 2}
	TxTagAppEcoShareTransfer = []byte{1, 2, 1}
	TxTagAppEcoGuarantee     = []byte{1, 2, 2}
	TxTagAppEcoRedeem        = []byte{1, 2, 3}
)

var (
	TxTagAngine = []byte{2}

	TxTagAngineEco        = []byte{2, 1}
	TxTagAngineEcoSuspect = []byte{2, 1, 1}

	TxTagAngineInit      = []byte{2, 2}
	TxTagAngineInitToken = []byte{2, 2, 1}
	TxTagAngineInitShare = []byte{2, 2, 2}

	TxTagAngineWorld     = []byte{2, 3}
	TxTagAngineWorldRand = []byte{2, 3, 1}
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
	GHeight       uint64
}

type WorldRandVote struct {
	Pubkey []byte
	Sig    []byte
}
