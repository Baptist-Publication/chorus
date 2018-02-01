package node

import (
	"bytes"
	"fmt"
	"math/big"

	agtypes "github.com/Baptist-Publication/angine/types"
	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus-module/xlib/def"
	cvtools "github.com/Baptist-Publication/chorus/src/tools"
	cvtypes "github.com/Baptist-Publication/chorus/src/types"
	"github.com/pkg/errors"
)

var (
	EcoTag           = []byte{'e', 'c', 'o'}
	EcoMortgageTag   = append(EcoTag, 0x01)
	EcoRedemptionTag = append(EcoTag, 0x02)
	EcoTransferTag   = append(EcoTag, 0x03)
	EcoInitAllocTag  = append(EcoTag, 0x04)
)

type EcoMortgageTx struct {
	cvtypes.CivilTx

	Amount *big.Int `json:"amount"`
}

type EcoRedemptionTx struct {
	cvtypes.CivilTx

	Amount *big.Int `json:"amount"`
}

type EcoTransferTx struct {
	cvtypes.CivilTx

	To     []byte   `json:"to"`
	Amount *big.Int `json:"amount"`
}

type EcoInitAllocTx struct {
	To     []byte   `json:"to"`
	Amount *big.Int `json:"amount"`
	Extra  []byte   `json:"extra"`
}

func IsEcoTx(tx []byte) bool {
	return bytes.HasPrefix(tx, EcoTag)
}

func IsEcoMortgageTx(tx []byte) bool {
	return bytes.HasPrefix(tx, EcoMortgageTag)
}

func IsEcoRedemptionTx(tx []byte) bool {
	return bytes.HasPrefix(tx, EcoRedemptionTag)
}

func IsEcoTransferTx(tx []byte) bool {
	return bytes.HasPrefix(tx, EcoTransferTag)
}

func IsEcoInitAllocTx(tx []byte) bool {
	return bytes.HasPrefix(tx, EcoInitAllocTag)
}

func (met *Metropolis) executeEcoInitAllocTx(tx *EcoInitAllocTx) error {
	var to crypto.PubKeyEd25519
	copy(to[:], tx.To)

	if err := met.accState.AddBalance(&to, tx.Amount); err != nil {
		return err
	}

	fmt.Printf("============ %X Init Alloc %s\n", tx.To, tx.Amount.String())

	return nil
}

func (met *Metropolis) executeEcoMortgageTx(tx *EcoMortgageTx, height def.INT) error {
	var from crypto.PubKeyEd25519
	copy(from[:], tx.GetPubKey())

	if tx.Amount.Cmp(cvtypes.BigInt0()) < 0 {
		return errors.New("amount cannot be negtive")
	}

	if err := met.accState.AssetNonce(tx.GetPubKey(), tx.GetNonce()); err != nil {
		return err
	}

	totalSub := new(big.Int).Add(tx.Amount, tx.Fee)
	if totalSub.Cmp(met.accState.QueryBalance(&from)) > 0 {
		return fmt.Errorf("insufficient balance")
	}

	if err := met.accState.SubBalance(&from, totalSub); err != nil {
		return err
	}

	if err := met.powerState.AddVTPower(&from, tx.Amount, height); err != nil {
		return err
	}

	if err := met.accState.IncreaseNonce(&from, 1); err != nil {
		return err
	}

	met.FeeAccum = new(big.Int).Add(met.FeeAccum, tx.Fee)

	fmt.Printf("============ %X Mortgage %s\n", tx.GetPubKey(), tx.Amount.String())

	return nil
}

func (met *Metropolis) executeEcoRedemptionTx(tx *EcoRedemptionTx, height def.INT) error {
	var from crypto.PubKeyEd25519
	copy(from[:], tx.GetPubKey())

	if height < def.StartRedeemHeight {
		return errors.New("Operation is not allow right now")
	}

	if tx.Amount.Cmp(big.NewInt(0)) < 0 {
		return errors.New("amount cannot be negtive")
	}

	if err := met.accState.AssetNonce(tx.GetPubKey(), tx.GetNonce()); err != nil {
		return err
	}

	if met.accState.QueryBalance(&from).Cmp(tx.Fee) < 0 {
		return fmt.Errorf("insufficient balance")
	}

	_, vSet := met.node.MainOrg.Angine.GetValidators()
	if vSet.HasAddress(agtypes.GetAddress(&from, met.core.GetChainID())) {
		// operator is validator now, cannot redeem right now, but we mark its MHeight as -1
		met.powerState.MarkPower(&from, -1)

		fmt.Printf("============ %X Redemption Mark\n", tx.GetPubKey())
	} else {
		if err := met.powerState.SubVTPower(&from, tx.Amount, height); err != nil {
			return err
		}
		if err := met.accState.AddBalance(&from, new(big.Int).Sub(tx.Amount, tx.Fee)); err != nil {
			return err
		}

		fmt.Printf("============ %X Redemption %s\n", tx.GetPubKey(), tx.Amount.String())
	}

	if err := met.accState.IncreaseNonce(&from, 1); err != nil {
		return err
	}

	met.FeeAccum = new(big.Int).Add(met.FeeAccum, tx.Fee)

	return nil
}

func (met *Metropolis) executeEcoTransferTx(tx *EcoTransferTx, height def.INT) error {
	var from, to crypto.PubKeyEd25519
	copy(from[:], tx.GetPubKey())
	copy(to[:], tx.To)

	if tx.Amount.Cmp(big.NewInt(0)) < 0 {
		return errors.New("amount cannot be negtive")
	}

	if err := met.accState.AssetNonce(tx.GetPubKey(), tx.GetNonce()); err != nil {
		return err
	}

	totalSub := new(big.Int).Add(tx.Amount, tx.Fee)
	if totalSub.Cmp(met.accState.QueryBalance(&from)) > 0 {
		return fmt.Errorf("insufficient balance")
	}

	if err := met.accState.SubBalance(&from, totalSub); err != nil {
		return err
	}

	if err := met.accState.AddBalance(&to, tx.Amount); err != nil {
		return err
	}

	if err := met.accState.IncreaseNonce(&from, 1); err != nil {
		return err
	}

	met.FeeAccum = new(big.Int).Add(met.FeeAccum, tx.Fee)

	fmt.Printf("============ %X Transfer %s\n", tx.GetPubKey(), tx.Amount.String())

	return nil
}

func (met *Metropolis) GetEcoCivilTx(bs []byte) (cvTx cvtypes.ICivilTx, err error) {
	txBytes := agtypes.UnwrapTx(bs)

	switch {
	case IsEcoMortgageTx(bs):
		tx := &EcoMortgageTx{}
		if err = cvtools.TxFromBytes(txBytes, tx); err != nil {
			return nil, errors.Wrap(err, "parse eco mortgage failed")
		}
		cvTx = tx
	case IsEcoRedemptionTx(bs):
		tx := &EcoRedemptionTx{}
		if err = cvtools.TxFromBytes(txBytes, tx); err != nil {
			return nil, errors.Wrap(err, "parse eco redemption failed")
		}
		cvTx = tx
	case IsEcoTransferTx(bs):
		tx := &EcoTransferTx{}
		if err = cvtools.TxFromBytes(txBytes, tx); err != nil {
			return nil, errors.Wrap(err, "parse eco transfer failed")
		}
		cvTx = tx
	}

	return
}
