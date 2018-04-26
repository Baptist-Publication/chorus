package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"

	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	cmn "github.com/Baptist-Publication/chorus/module/lib/go-common"
	crypto "github.com/Baptist-Publication/chorus/module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus/module/xlib/def"
	ethcmn "github.com/Baptist-Publication/chorus/eth/common"
	"github.com/Baptist-Publication/chorus/eth/core"
	ethtypes "github.com/Baptist-Publication/chorus/eth/core/types"
	"github.com/Baptist-Publication/chorus/eth/params"
	"github.com/Baptist-Publication/chorus/tools"
	"github.com/Baptist-Publication/chorus/types"
	ctypes "github.com/Baptist-Publication/chorus/types"
)

func (app *App) RegisterValidators(validatorset *agtypes.ValidatorSet) {
	for _, validator := range validatorset.Validators {
		if pub, ok := validator.GetPubKey().(*crypto.PubKeyEd25519); ok {
			// app.accState.CreateAccount(pub[:], Big0)
			app.currentShareState.AddGuaranty(pub, new(big.Int).SetUint64(uint64(validator.VotingPower)), 1)
		}
	}
}

type comparableShare Share

func (ca *comparableShare) Less(o interface{}) bool {
	return ca.ShareBalance.Cmp(o.(*comparableShare).ShareBalance) < 0
}

func CalcVP(base int, position int) uint64 {
	for level := 1; ; level++ {
		if position <= level*(level+1)/2 {
			return uint64((base - level + 1) * (base - level + 1))
		}
	}
}

func (app *App) ValSetLoader() agtypes.ValSetLoaderFunc {
	return func(height, round def.INT, size int) *agtypes.ValidatorSet {
		vals := app.fakeRandomVals(height, round, size)

		return agtypes.NewValidatorSet(vals)
	}
}

func (app *App) fakeRandomVals(height, round def.INT, size int) []*agtypes.Validator {
	pwrs := make([]*Share, 0, 21)

	// Iterate power list of world state
	// to get all power account than joining current election
	accList := make([]*Share, 0, app.ShareState.Size())
	vsetHeap := cmn.NewHeap() // max heap of length 15
	app.ShareState.Iterate(func(pwr *Share) bool {
		if pwr.MHeight == -1 { // indicate he is not in
			return false
		}

		accList = append(accList, pwr)
		vsetHeap.Push(pwr, (*comparableShare)(pwr))
		if vsetHeap.Len() > 15 {
			vsetHeap.Pop()
		}
		return false
	})

	// Pick all and share the same power
	// if the length of partners is less than 6
	if len(accList) <= 6 {
		vals := make([]*agtypes.Validator, len(accList))
		for i, v := range accList {
			var pk crypto.PubKeyEd25519
			copy(pk[:], v.Pubkey)
			vals[i] = &agtypes.Validator{PubKey: crypto.StPubKey{PubKey: &pk}, Address: pk.Address(), VotingPower: 100}
		}
		return vals
	}

	// Calculate the number of power account that need to select 'randomly' from the rest account
	// n = len(accList)
	//  if      n <= 15 	then 0
	//  if 15 < n <= 21 	then 15 - n
	//  if 21 < n			then 6
	numLuckyGuys := int(len(accList)) - 15
	if numLuckyGuys < 0 {
		numLuckyGuys = 0
	} else if numLuckyGuys > 6 {
		numLuckyGuys = 6
	}

	// Pick the rich-guys
	// max(rich-guys) = 15
	// min(rich-guys) = 7
	exists := make(map[*Share]struct{})
	i := int(vsetHeap.Len())
	for vsetHeap.Len() > 0 {
		pwr := vsetHeap.Pop().(*Share)
		pwr.ShareBalance = new(big.Int).SetUint64(CalcVP(6, i)) // re-calculate power
		pwrs = append(pwrs, pwr)
		exists[pwr] = struct{}{}
		i--
	}

	// Pick lucky-guys 'randomly' from the rest of power account
	// we use a map(means exists) to identify the elected guys
	retry := 1
	bigbang := new(big.Int).SetBytes(app.evmState.IntermediateRoot(true).Bytes())
	luckyguys := make([]*Share, 0, numLuckyGuys)
	for len(luckyguys) < numLuckyGuys {
		guy, err := fakeRandomAccount(accList, exists, height, round, bigbang, &retry)
		if err != nil {
			fmt.Println("error in fakeRandomAccount:", err.Error())
			return nil
		}
		guy.ShareBalance = new(big.Int).SetUint64(1)
		luckyguys = append(luckyguys, guy)
	}

	// combine ...
	pwrs = append(pwrs, luckyguys...)

	// make validators according to the power accounts elected above
	vals := make([]*agtypes.Validator, len(pwrs))
	for i, v := range pwrs {
		var pk crypto.PubKeyEd25519
		copy(pk[:], v.Pubkey)
		vals[i] = &agtypes.Validator{
			PubKey:      crypto.StPubKey{PubKey: &pk},
			Address:     pk.Address(),
			VotingPower: v.ShareBalance.Int64(),
		}
	}

	return vals
}

func fakeRandomAccount(accs []*Share, exists map[*Share]struct{}, height, round def.INT, bigbang *big.Int, retry *int) (*Share, error) {
	if len(accs) == len(exists) {
		return nil, fmt.Errorf("No account can be picked any more")
	}

	base := new(big.Int).Add(bigbang, new(big.Int).SetUint64(uint64(height+1)*uint64(*retry+1)))
	index := new(big.Int).Mod(base, new(big.Int).SetUint64(uint64(len(accs)))).Int64()
	for {
		(*retry)++
		tryPick := accs[index]
		if _, yes := exists[tryPick]; !yes {
			exists[tryPick] = struct{}{}
			return tryPick, nil
		}
		if index == int64(len(accs))-1 {
			index = 0
		} else {
			index++
		}
	}
}

func (app *App) CheckEcoTx(bs []byte) error {
	// nothing to do
	return nil
}

//doCoinbaseTx send block rewards to block maker
func (app *App) doCoinbaseTx(block *agtypes.BlockCache) error {
	var addr ethcmn.Address
	copy(addr[:], block.Header.CoinBase)

	rewards := calculateRewards(uint64(block.Header.Height))

	app.currentEvmState.AddBalance(addr, rewards)
	return nil
}

func calculateRewards(height uint64) *big.Int {
	startRewards := uint64(128 * 1000000000)
	declinePerBlocks := uint64(5000000)

	declineCount := height / declinePerBlocks
	if declineCount > 7 {
		declineCount = 7
	}
	declineBase := uint64(math.Pow(2, float64(declineCount)))
	rewards := startRewards / declineBase

	return new(big.Int).SetUint64(rewards)
}

func (app *App) ExecuteAppEcoTx(block *agtypes.BlockCache, bs []byte, tx *ctypes.BlockTx) (hash []byte, usedGas *big.Int, err error) {
	switch {
	case bytes.HasPrefix(bs, types.TxTagAppEcoShareTransfer):
		err = app.executeShareTransfer(block, tx)
	case bytes.HasPrefix(bs, types.TxTagAppEcoGuarantee):
		err = app.executeShareGuarantee(block, tx)
	case bytes.HasPrefix(bs, types.TxTagAppEcoRedeem):
		err = app.executeShareRedeem(block, tx)
	}
	return nil, big0, err
}

func (app *App) ExecuteAppInitTx(block *agtypes.BlockCache, bs []byte, txIndex int) (hash []byte, usedGas *big.Int, err error) {
	if app.AngineRef.Height() != 1 {
		return nil, nil, fmt.Errorf("Insufficient block height")
	}

	switch {
	case bytes.HasPrefix(bs, types.TxTagAngineInitToken):
		err = app.executeTokenInitTx(agtypes.UnwrapTx(bs))
	case bytes.HasPrefix(bs, types.TxTagAngineInitShare):
		err = app.executeShareInitTx(agtypes.UnwrapTx(bs))
	}

	return nil, big0, err
}

func (app *App) executeTokenInitTx(bs []byte) error {
	var tx ctypes.EcoInitTokenTx
	err := json.Unmarshal(bs, &tx)
	if err != nil {
		return fmt.Errorf("Unmarshal tx failed:%s", err.Error())
	}

	var addr ethcmn.Address
	copy(addr[:], tx.To)
	app.currentEvmState.AddBalance(addr, tx.Amount)

	return nil
}

func (app *App) executeShareInitTx(bs []byte) error {
	var tx ctypes.EcoInitShareTx
	err := json.Unmarshal(bs, &tx)
	if err != nil {
		return fmt.Errorf("Unmarshal tx failed:%s", err.Error())
	}

	app.currentShareState.CreateShareAccount(tx.To, tx.Amount, 1)

	return nil
}

func (app *App) executeShareTransfer(block *agtypes.BlockCache, tx *ctypes.BlockTx) error {
	bodytx := ctypes.TxShareTransfer{}
	if err := tools.TxFromBytes(tx.Payload, &bodytx); err != nil {
		return err
	}
	//VerifySignature
	if right, err := bodytx.VerifySig(); err != nil || !right {
		return errors.New("verify signatrue failed")
	}
	if err := app.chargeFee(block, tx); err != nil {
		return err
	}
	//transfer
	frompub, topub := crypto.PubKeyEd25519{}, crypto.PubKeyEd25519{}
	copy(frompub[:], bodytx.ShareSrc)
	copy(topub[:], bodytx.ShareDst)
	err := app.currentShareState.SubShareBalance(&frompub, bodytx.Amount)
	if err != nil {
		return err
	}
	app.currentShareState.AddShareBalance(&topub, bodytx.Amount, block.Header.Height)
	//save receipts
	app.addReceipt(tx)
	return nil
}

func (app *App) executeShareGuarantee(block *agtypes.BlockCache, tx *ctypes.BlockTx) error {
	bodytx := ctypes.TxShareEco{}
	if err := tools.TxFromBytes(tx.Payload, &bodytx); err != nil {
		return err
	}
	//VerifySignature
	if right, err := bodytx.VerifySig(); err != nil || !right {
		return errors.New("verify signatrue failed")
	}
	if err := app.chargeFee(block, tx); err != nil {
		return err
	}
	//transfer
	frompub := crypto.PubKeyEd25519{}
	copy(frompub[:], bodytx.Source)
	err := app.currentShareState.SubShareBalance(&frompub, bodytx.Amount)
	if err != nil {
		return err
	}
	app.currentShareState.AddGuaranty(&frompub, bodytx.Amount, block.Header.Height)
	//save receipts
	app.addReceipt(tx)
	return nil
}

func (app *App) executeShareRedeem(block *agtypes.BlockCache, tx *ctypes.BlockTx) error {
	bodytx := ctypes.TxShareEco{}
	if err := tools.TxFromBytes(tx.Payload, &bodytx); err != nil {
		return err
	}
	//VerifySignature
	if right, err := bodytx.VerifySig(); err != nil || !right {
		return errors.New("verify signatrue failed")
	}
	if err := app.chargeFee(block, tx); err != nil {
		return err
	}

	frompub := crypto.PubKeyEd25519{}
	copy(frompub[:], bodytx.Source)
	err := app.currentShareState.SubGuaranty(&frompub, bodytx.Amount)
	if err != nil {
		return err
	}
	app.currentShareState.AddShareBalance(&frompub, bodytx.Amount, block.Header.Height)
	//save receipts
	app.addReceipt(tx)
	return nil
}

func (app *App) chargeFee(block *agtypes.BlockCache, tx *ctypes.BlockTx) error {
	sender := ethcmn.BytesToAddress(tx.Sender)
	//check nonce
	if n := app.currentEvmState.GetNonce(sender); n != tx.Nonce {
		return core.NonceError(tx.Nonce, n)
	}
	//check balance
	mval := new(big.Int).Mul(tx.GasLimit, tx.GasPrice)
	balance := app.currentEvmState.GetBalance(sender)
	if balance.Cmp(mval) < 0 {
		return fmt.Errorf("insufficient balance for gas (%x). Req %v, has %v", sender.Bytes()[:4], mval, balance)
	}
	if tx.GasLimit.Cmp(params.TxGas) < 0 {
		return errors.New("out of gas")
	}
	//do charge
	realfee := new(big.Int).Mul(params.TxGas, tx.GasPrice)
	app.currentEvmState.SubBalance(sender, realfee)
	app.currentEvmState.AddBalance(ethcmn.BytesToAddress(block.Header.CoinBase), realfee)
	//set nonce
	app.currentEvmState.SetNonce(sender, app.currentEvmState.GetNonce(sender)+1)
	return nil
}

func (app *App) addReceipt(tx *ctypes.BlockTx) {
	receipt := ethtypes.NewReceipt([]byte{}, params.TxGas)
	receipt.TxHash = ethcmn.BytesToHash(tx.Hash())
	receipt.GasUsed = params.TxGas
	receipt.Logs = app.currentEvmState.GetLogs(ethcmn.BytesToHash(tx.Hash()))
	receipt.Bloom = ethtypes.CreateBloom(ethtypes.Receipts{receipt})
	app.receipts = append(app.receipts, receipt)
}
