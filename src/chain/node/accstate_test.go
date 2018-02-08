package node

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	cmn "github.com/Baptist-Publication/chorus-module/lib/go-common"
	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	"github.com/tendermint/tmlibs/db"
)

func getAccState() *AccState {
	accDB, err := db.NewGoLevelDB(fmt.Sprintf("accountstate%d", rand.Int()), "/tmp/")
	if err != nil {
		panic(err)
	}

	return NewAccState(accDB)
}

func getPowerState() *PowerState {
	accDB, err := db.NewGoLevelDB(fmt.Sprintf("powerstate%d", rand.Int()), "/tmp/")
	if err != nil {
		panic(err)
	}

	return NewPowerState(accDB)
}

func randomPubkey() (*crypto.PubKeyEd25519, error) {
	sk1, err := crypto.GenPrivKeyEd25519(nil)
	if err != nil {
		return nil, err
	}
	defer sk1.Destroy()
	pk2 := sk1.PubKey()
	pk2bs := pk2.(*crypto.PubKeyEd25519)

	return pk2bs, nil
}

func randomBig(max uint64) *big.Int {
	return new(big.Int).SetUint64(rand.Uint64() % max)
}

func _checkErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestIterateAccountState(t *testing.T) {
	as := getAccState()

	var k1, k2, k3, k4 *crypto.PubKeyEd25519
	var err error
	k1, err = randomPubkey()
	_checkErr(t, err)
	k2, err = randomPubkey()
	_checkErr(t, err)
	k3, err = randomPubkey()
	_checkErr(t, err)
	k4, err = randomPubkey()
	_checkErr(t, err)

	as.AddBalance(k1, new(big.Int).SetUint64(10))
	as.AddBalance(k2, new(big.Int).SetUint64(20))
	as.AddBalance(k3, new(big.Int).SetUint64(30))
	as.AddBalance(k4, new(big.Int).SetUint64(40))

	root, _ := as.Commit()
	as.Reload(root)

	as.Iterate(func(acc *Account) bool {
		fmt.Println(acc.Balance.String())
		return false
	})
	fmt.Println("====================================")

	as.SubBalance(k2, new(big.Int).SetUint64(20))
	as.SubBalance(k3, new(big.Int).SetUint64(5))

	as.Iterate(func(acc *Account) bool {
		fmt.Println(acc.Balance.String())
		return false
	})
}

func TestIteratePowerState(t *testing.T) {
	ps := getPowerState()

	var k1, k2, k3, k4 *crypto.PubKeyEd25519
	var err error
	k1, err = randomPubkey()
	_checkErr(t, err)
	k2, err = randomPubkey()
	_checkErr(t, err)
	k3, err = randomPubkey()
	_checkErr(t, err)
	k4, err = randomPubkey()
	_checkErr(t, err)

	ps.CreatePower(k1[:], new(big.Int).SetUint64(10), 1)
	ps.CreatePower(k2[:], new(big.Int).SetUint64(20), 1)
	ps.CreatePower(k3[:], new(big.Int).SetUint64(30), 1)
	ps.CreatePower(k4[:], new(big.Int).SetUint64(40), 1)

	root, _ := ps.Commit()
	ps.Reload(root)

	ps.Iterate(func(pwr *Power) bool {
		fmt.Println(pwr.VTPower.String())
		return false
	})
	fmt.Println("====================================")

	ps.SubVTPower(k2, new(big.Int).SetUint64(20), 2)
	ps.SubVTPower(k3, new(big.Int).SetUint64(5), 2)

	ps.Iterate(func(pwr *Power) bool {
		fmt.Println(pwr.VTPower.String())
		return false
	})
}

func TestElection(t *testing.T) {
	ps := getPowerState()
	as := getAccState()
	met := &Metropolis{
		powerState: ps,
		accState:   as,
	}

	for i := 0; i < 50; i++ {
		k, err := randomPubkey()
		_checkErr(t, err)
		p := new(big.Int).SetUint64(uint64(10 + i*10))
		ps.CreatePower(k[:], p, 1)
		as.CreateAccount(k[:], Big0)
		fmt.Printf("%X-%s\n", k[:2], p.String())
	}

	root, _ := ps.Commit()
	ps.Reload(root)
	root2, _ := as.Commit()
	as.Reload(root2)

	// fmt.Println("Accounts in world state:", ps.trie.Size())

	vals := met.fakeRandomVals(1, 1, 21)
	fmt.Println(len(vals))
	for _, v := range vals {
		fmt.Printf("%s-%d ", v.PubKey.KeyString()[:4], v.VotingPower)
	}
	fmt.Println()

	// vs := met.ValSetLoader()(1, 1, 21)
	// for t := 0; t < 10; t++ {
	// 	for _, v := range vs.Validators {
	// 		fmt.Printf("%8d", v.Accum)
	// 	}
	// 	fmt.Printf("  proposer: %-8d", vs.Proposer().PubKey)
	// 	fmt.Println()
	// 	vs.IncrementAccum(1)
	// 	for _, v := range vs.Validators {
	// 		fmt.Printf("%X %d %d\n", v.PubKey.KeyString()[:4], v.VotingPower, v.Accum)
	// 	}
	// }
}

type compInt int

func (c compInt) Less(o interface{}) bool {
	return c < o.(compInt)
}

func TestNothing(t *testing.T) {
	vsetHeap := cmn.NewHeap()

	for i := 0; i < 10; i++ {
		v := rand.Int() % 1000
		vsetHeap.Push(v, compInt(v))
	}

	for vsetHeap.Len() > 0 {
		fmt.Println(vsetHeap.Pop())
	}
}
