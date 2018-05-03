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

		vals := app.doElect(new(big.Int).SetBytes(worldRand), height, round)
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

func (app *App) doElect(bigbang *big.Int, height, round def.INT) []*agtypes.Validator {
	shrs := make([]*Share, 0, 21)

	// Iterate share list of world state
	// to get all share accounts that taking part in current election
	exists := make(map[*Share]struct{})
	accList := make([]*Share, 0, app.ShareState.Size())
	accHeap := cmn.NewHeap() // max-heap
	app.ShareState.Iterate(func(shr *Share) bool {
		if shr.GHeight == 0 || shr.ShareGuaranty.Cmp(big0) == 0 { // indicate he is not in
			return false
		}

		newShr := shr.Copy()
		accList = append(accList, newShr)
		accHeap.Push(newShr, newShr)
		if accHeap.Len() > 32 {
			accHeap.Pop()
		}
		return false
	})

	// Pick the rich-guys
	richGuys := make([]*Share, 0, accHeap.Len())
	for accHeap.Len() > 0 {
		shr := accHeap.Pop().(*Share)
		richGuys = append(richGuys, shr)
	}
	pickedRichGuys, err := pickAccounts(richGuys, exists, 11, bigbang)
	if err != nil {
		fmt.Println("error in pickAccounts:", err.Error())
		return nil
	}
	shrs = append(shrs, pickedRichGuys...)

	// Pick lucky-guys
	//  if      n <= 11 	then 0
	//  if 11 < n <= 15 	then n - 11
	//  if 15 < n			then 4
	numLuckyGuys := len(accList) - 11
	if numLuckyGuys < 0 {
		numLuckyGuys = 0
	} else if numLuckyGuys > 4 {
		numLuckyGuys = 4
	}
	luckyguys, err := pickAccounts(accList, exists, numLuckyGuys, bigbang)
	if err != nil {
		fmt.Println("error in pickAccounts:", err.Error())
		return nil
	}

	// combine ...
	shrs = append(shrs, luckyguys...)
	// make validators according to the power accounts elected above
	vals := make([]*agtypes.Validator, len(shrs))
	for i, v := range shrs {
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

func pickAccounts(froms []*Share, exists map[*Share]struct{}, maxPick int, bigbang *big.Int) ([]*Share, error) {
	if maxPick == 0 {
		return make([]*Share, 0), nil
	}

	if len(froms) <= maxPick {
		return froms, nil
	}

	guys := make([]*Share, 0, maxPick)
	for len(guys) < maxPick {
		guy, err := randomSelect(froms, exists, bigbang)
		if err != nil {
			fmt.Println("error in randomSelect:", err.Error())
			return nil, nil
		}
		guys = append(guys, guy)
	}

	return guys, nil
}

func randomSelect(accs []*Share, exists map[*Share]struct{}, bigbang *big.Int) (*Share, error) {
	if len(accs) == len(exists) {
		return nil, fmt.Errorf("No account can be picked any more")
	}

	type shareVotes struct {
		ref        *Share
		votesBegin *big.Int
		votesEnd   *big.Int
	}

	totalShare := new(big.Int)
	votes := make([]shareVotes, 0, len(accs))
	for i := 0; i < len(accs); i++ {
		if _, yes := exists[accs[i]]; !yes {
			votes = append(votes, shareVotes{
				ref:        accs[i],
				votesBegin: new(big.Int).Set(totalShare),
				votesEnd:   new(big.Int).Add(totalShare, accs[i].ShareGuaranty),
			})
			totalShare.Add(totalShare, accs[i].ShareGuaranty)
		}
	}

	lottery := new(big.Int).Mod(bigbang, totalShare)
	for i := 0; i < len(votes); i++ {
		if lottery.Cmp(votes[i].votesEnd) < 0 {
			exists[votes[i].ref] = struct{}{}
			return votes[i].ref, nil
		}
	}

	fmt.Println("Should never happen")
	return nil, nil
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

	app.currentShareState.CreateShareAccount(tx.To, tx.Amount)

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
	app.currentShareState.AddShareBalance(&topub, bodytx.Amount)
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
	//check ElectionTerm
	height := uint64(app.AngineRef.Height())
	begin, _, mid := app.calElectionTerm(height)
	if height < begin || height >= mid {
		return fmt.Errorf("Guarantee tx happenes in invalid height(%d). current term:[%d,%d]", height, begin, mid-1)
	}

	//do guarantee
	frompub := crypto.PubKeyEd25519{}
	copy(frompub[:], bodytx.Source)
	err := app.currentShareState.SubShareBalance(&frompub, bodytx.Amount)
	if err != nil {
		return err
	}
	app.currentShareState.AddGuaranty(&frompub, bodytx.Amount, uint64(block.Header.Height))
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
	//check ElectionTerm
	height := uint64(app.AngineRef.Height())
	begin, _, mid := app.calElectionTerm(height)
	if height < begin || height >= mid {
		return fmt.Errorf("Redeem tx happenes in invalid height(%d). current term:[%d,%d]", height, begin, mid-1)
	}

	//do redeem
	frompub := crypto.PubKeyEd25519{}
	copy(frompub[:], bodytx.Source)
	// if node is validator right now
	if app.AngineRef.IsNodeValidator(&frompub) {
		app.currentShareState.MarkShare(&frompub, 0)
	} else {
		if err := app.currentShareState.SubGuaranty(&frompub, bodytx.Amount); err != nil {
			return err
		}
		if err := app.currentShareState.AddShareBalance(&frompub, bodytx.Amount); err != nil {
			return err
		}
	}

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
