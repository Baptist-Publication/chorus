package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/vmihailenco/msgpack"

	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	ethcmn "github.com/Baptist-Publication/chorus/eth/common"
	"github.com/Baptist-Publication/chorus/eth/core"
	ethtypes "github.com/Baptist-Publication/chorus/eth/core/types"
	"github.com/Baptist-Publication/chorus/eth/params"
	cmn "github.com/Baptist-Publication/chorus/module/lib/go-common"
	crypto "github.com/Baptist-Publication/chorus/module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus/module/lib/go-merkle"
	"github.com/Baptist-Publication/chorus/module/xlib/def"
	"github.com/Baptist-Publication/chorus/tools"
	"github.com/Baptist-Publication/chorus/types"
	"go.uber.org/zap"
)

const (
	WorldRandPrefix = "world-rand-"
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
	election := app.Config.GetInt64("election")

	return func(height, round def.INT) *agtypes.ValidatorSet {
		if height%election > 0 {
			return nil
		}

		worldRand, err := app.getWorldRand(uint64(height))
		if err != nil {
			return nil
		}
		fmt.Printf("======= Finally we got world rand : %x\n", worldRand)

		vals := app.fakeRandomVals(new(big.Int).SetBytes(worldRand), height, round)
		return agtypes.NewValidatorSet(vals)
	}
}

func (app *App) getWorldRand(height uint64) ([]byte, error) {
	// calc global rand
	var votes []types.WorldRandVote
	begin, _, _ := app.calElectionTerm(height)
	key := makeWorldRandKey(begin)
	value := app.Database.Get(key)
	if len(value) > 0 {
		if err := msgpack.Unmarshal(value, &votes); err != nil {
			app.logger.Error("Unmarshal globalrand votes failed", zap.Error(err))
			return nil, err
		}
	}
	fmt.Printf("======= We got %d votes for height %d\n", len(votes), begin)

	votesAmt := app.Config.GetInt("worldrand_votes_amount")
	bs := make([][]byte, votesAmt)
	for i := 0; i < votesAmt && i < len(votes); i++ {
		bs[i] = votes[i].Sig
	}
	worldRand := merkle.SimpleHashFromHashes(bs)

	return worldRand, nil
}

func (app *App) fakeRandomVals(bigbang *big.Int, height, round def.INT) []*agtypes.Validator {
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
	// bigbang := new(big.Int).SetBytes(app.evmState.IntermediateRoot(true).Bytes())
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

func (app *App) calElectionTerm(height uint64) (begin, end, mid uint64) {
	electionSpan := uint64(app.Config.GetInt64("election"))
	worldRandThreshold := app.Config.GetFloat64("worldrand_threshold")

	if height%electionSpan == 0 {
		height = height - 1
	}

	begin = height - (height % electionSpan) + 1
	end = begin + electionSpan - 1
	mid = begin + uint64(float64(electionSpan)*worldRandThreshold)

	return
}

func (app *App) prepareGlobalRand(height, round uint64) {
	privkey := app.AngineRef.PrivValidator().GetPrivKey()
	pubkey := app.AngineRef.PrivValidator().GetPubKey()

	begin, end, mid := app.calElectionTerm(uint64(height))
	if height < mid || height > end || round > 1 {
		return
	}
	if !app.AngineRef.IsNodeValidator(pubkey) {
		return
	}

	// hash privkey to get a fake random number
	fr := crypto.Sha256(privkey.Bytes())
	frv := new(big.Int).SetBytes(fr)
	frv = frv.Add(frv, new(big.Int).SetUint64((begin+1)*17))

	// calc when should we fire the GlobalRand tx
	target := mid + frv.Mod(frv, big.NewInt(int64(end-mid-1))).Uint64()
	if target != height {
		return
	}

	// Fire
	blockMeta, err := app.AngineRef.GetBlockMeta(int64(begin))
	if err != nil {
		app.logger.Error("GetBlockMeta failed:", zap.Error(err))
		return
	}
	sig := privkey.Sign(blockMeta.GetHash())

	tx := &types.WorldRandTx{
		Height: begin,
		Pubkey: pubkey.Bytes(),
		Sig:    sig.Bytes(),
	}
	txBytes, err := tools.TxToBytes(tx)
	if err != nil {
		app.logger.Error("TxToBytes failed:", zap.Error(err))
		return
	}
	b := agtypes.WrapTx(types.TxTagAngineWorldRand, txBytes)
	app.AngineRef.BroadcastTx(b)
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

func (app *App) ExecuteAppEcoTx(block *agtypes.BlockCache, bs []byte, tx *types.BlockTx) (hash []byte, usedGas *big.Int, err error) {
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

func (app *App) ExecuteInitTx(block *agtypes.BlockCache, bs []byte, txIndex int) (hash []byte, usedGas *big.Int, err error) {
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

func (app *App) ExecuteAngineWorldTx(block *agtypes.BlockCache, bs []byte, txIndex int) (hash []byte, usedGas *big.Int, err error) {
	switch {
	case bytes.HasPrefix(bs, types.TxTagAngineWorldRand):
		err = app.executeWorldRandTx(agtypes.UnwrapTx(bs))
	}

	if err != nil {
		app.logger.Error("Execute AngineWorldTx failed", zap.Error(err))
	}

	return nil, big0, err
}

func (app *App) executeWorldRandTx(bs []byte) error {
	height := uint64(app.AngineRef.Height())
	begin, end, mid := app.calElectionTerm(height)

	// check allowed height span
	if height < mid || height > end {
		return fmt.Errorf("WorldRandTx happens in invalid height: %d", height)
	}

	// decode
	var tx types.WorldRandTx
	err := tools.TxFromBytes(bs, &tx)
	if err != nil {
		return fmt.Errorf("TxFromBytes failed:%s", err.Error())
	}

	// check tx
	if tx.Height != begin {
		return fmt.Errorf("WorldRandTx has an invalid height: %d, expected: %d", tx.Height, begin)
	}
	pubkey := &crypto.PubKeyEd25519{}
	copy(pubkey[:], tx.Pubkey)
	if !app.AngineRef.IsNodeValidator(pubkey) {
		return fmt.Errorf("WorldRandTx was fired by non-validator node")
	}
	blockMeta, err := app.AngineRef.GetBlockMeta(int64(tx.Height))
	if err != nil {
		return err
	}
	blockHash := blockMeta.GetHash()
	sig := &crypto.SignatureEd25519{}
	copy(sig[:], tx.Sig)
	if !pubkey.VerifyBytes(blockHash, sig) {
		return fmt.Errorf("WorldRandTx has a invalid sig")
	}

	// save vote
	var votes []types.WorldRandVote
	key := makeWorldRandKey(begin)
	votesBs := app.Database.Get(key)
	if len(votesBs) > 0 {
		if err := msgpack.Unmarshal(votesBs, &votes); err != nil {
			return err
		}
		for _, v := range votes {
			if bytes.Equal(tx.Pubkey, v.Pubkey) {
				return nil // duplicated, nothing to do
			}
		}
	}
	votes = append(votes, types.WorldRandVote{
		Pubkey: tx.Pubkey,
		Sig:    tx.Sig,
	})
	newVotesBs, err := msgpack.Marshal(votes)
	if err != nil {
		return err
	}
	app.Database.SetSync(key, newVotesBs)

	fmt.Printf("Node %x vote for worldrand: %d in height %d\n", tx.Pubkey[:8], begin, height)
	return nil
}

func makeWorldRandKey(height uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", WorldRandPrefix, height))
}

func (app *App) executeTokenInitTx(bs []byte) error {
	var tx types.EcoInitTokenTx
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
	var tx types.EcoInitShareTx
	err := json.Unmarshal(bs, &tx)
	if err != nil {
		return fmt.Errorf("Unmarshal tx failed:%s", err.Error())
	}

	app.currentShareState.CreateShareAccount(tx.To, tx.Amount, 1)

	return nil
}

func (app *App) executeShareTransfer(block *agtypes.BlockCache, tx *types.BlockTx) error {
	bodytx := types.TxShareTransfer{}
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

func (app *App) executeShareGuarantee(block *agtypes.BlockCache, tx *types.BlockTx) error {
	bodytx := types.TxShareEco{}
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

func (app *App) executeShareRedeem(block *agtypes.BlockCache, tx *types.BlockTx) error {
	bodytx := types.TxShareEco{}
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

func (app *App) chargeFee(block *agtypes.BlockCache, tx *types.BlockTx) error {
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

func (app *App) addReceipt(tx *types.BlockTx) {
	receipt := ethtypes.NewReceipt([]byte{}, params.TxGas)
	receipt.TxHash = ethcmn.BytesToHash(tx.Hash())
	receipt.GasUsed = params.TxGas
	receipt.Logs = app.currentEvmState.GetLogs(ethcmn.BytesToHash(tx.Hash()))
	receipt.Bloom = ethtypes.CreateBloom(ethtypes.Receipts{receipt})
	app.receipts = append(app.receipts, receipt)
}
