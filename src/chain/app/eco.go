package app

import (
	"fmt"
	"math/big"

	agtypes "github.com/Baptist-Publication/angine/types"
	cmn "github.com/Baptist-Publication/chorus-module/lib/go-common"
	crypto "github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus-module/xlib/def"
)

func (app *App) RegisterValidators(validatorset *agtypes.ValidatorSet) {
	for _, validator := range validatorset.Validators {
		if pub, ok := validator.GetPubKey().(*crypto.PubKeyEd25519); ok {
			// app.accState.CreateAccount(pub[:], Big0)
			app.currentPowerState.CreatePower(pub[:], new(big.Int).SetUint64(uint64(validator.VotingPower)), 1)
		}
	}
}

type comparablePower Power

func (ca *comparablePower) Less(o interface{}) bool {
	return ca.VTPower.Cmp(o.(*comparablePower).VTPower) < 0
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
	pwrs := make([]*Power, 0, 21)

	// Iterate power list of world state
	// to get all power account than joining current election
	accList := make([]*Power, 0, app.powerState.Size())
	vsetHeap := cmn.NewHeap() // max heap of length 15
	app.powerState.Iterate(func(pwr *Power) bool {
		if pwr.MHeight == -1 { // indicate he is not in
			return false
		}

		accList = append(accList, pwr)
		vsetHeap.Push(pwr, (*comparablePower)(pwr))
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
	exists := make(map[*Power]struct{})
	i := int(vsetHeap.Len())
	for vsetHeap.Len() > 0 {
		pwr := vsetHeap.Pop().(*Power)
		pwr.VTPower = new(big.Int).SetUint64(CalcVP(6, i)) // re-calculate power
		pwrs = append(pwrs, pwr)
		exists[pwr] = struct{}{}
		i--
	}

	// Pick lucky-guys 'randomly' from the rest of power account
	// we use a map(means exists) to identify the elected guys
	retry := 1
	bigbang := new(big.Int).SetBytes(app.evmState.IntermediateRoot(true).Bytes())
	luckyguys := make([]*Power, 0, numLuckyGuys)
	for len(luckyguys) < numLuckyGuys {
		guy, err := fakeRandomAccount(accList, exists, height, round, bigbang, &retry)
		if err != nil {
			fmt.Println("error in fakeRandomAccount:", err.Error())
			return nil
		}
		guy.VTPower = new(big.Int).SetUint64(1)
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
			VotingPower: v.VTPower.Int64(),
		}
	}

	return vals
}

func fakeRandomAccount(accs []*Power, exists map[*Power]struct{}, height, round def.INT, bigbang *big.Int, retry *int) (*Power, error) {
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
