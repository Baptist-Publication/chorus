package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/tendermint/tmlibs/db"
	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus-module/xlib/iavl"
	"github.com/Baptist-Publication/chorus-module/xlib/mlist"
)

type AccState struct {
	mtx      sync.Mutex
	database db.DB
	rootHash []byte
	trie     *iavl.IAVLTree

	//key is ed25519 pubkey
	accountCache *mlist.MapList
}

type Account struct {
	Nonce   uint64
	Pubkey  []byte
	Balance *big.Int
}

func NewAccState(database db.DB) *AccState {
	return &AccState{
		//dirty:        make(map[string]struct{}),
		database:     database,
		trie:         iavl.NewIAVLTree(1024, database),
		accountCache: mlist.NewMapList(),
	}
}

func (as *AccState) Lock() {
	as.mtx.Lock()
}

func (as *AccState) Unlock() {
	as.mtx.Unlock()
}

func (as *AccState) CreateAccount(pubkey []byte, balance *big.Int) {
	if balance.Cmp(big.NewInt(0)) < 0 {
		return
	}
	as.mtx.Lock()
	defer as.mtx.Unlock()

	pub := crypto.PubKeyEd25519{}
	copy(pub[:], pubkey[:])

	acc := &Account{
		Pubkey:  pubkey,
		Balance: new(big.Int).Set(balance),
		Nonce:   1,
	}

	as.accountCache.Set(pub.KeyString(), acc)
}

func (as *AccState) GetAccount(pubkey []byte) (*Account, error) {
	as.mtx.Lock()
	defer as.mtx.Unlock()
	pub := crypto.PubKeyEd25519{}
	copy(pub[:], pubkey)

	if acc, ok := as.accountCache.Get(pub.KeyString()); ok {
		return acc.(*Account), nil
	}
	if _, accountbytes, exist := as.trie.Get([]byte(pub.KeyString())); exist {
		acc := new(Account)
		acc.FromBytes(accountbytes)
		return acc, nil
	}
	return nil, fmt.Errorf("account not exist: %X", pubkey)
}

func (as *AccState) AddBalance(pubkey crypto.PubKey, amount *big.Int) error {
	keystring := pubkey.KeyString()
	as.Lock()
	defer as.Unlock()

	// from cache
	if itfc, ok := as.accountCache.Get(keystring); ok {
		acc := itfc.(*Account)
		acc.Balance = new(big.Int).Add(acc.Balance, amount)
		as.accountCache.Set(keystring, acc)
		return nil
	}

	// from db
	if _, value, exist := as.trie.Get([]byte(keystring)); exist {
		acc := new(Account)
		err := acc.FromBytes(value)
		if err != nil {
			return err
		}
		acc.Balance = new(big.Int).Add(acc.Balance, amount)
		as.accountCache.Set(keystring, acc)
		return nil
	}

	// new account
	pk := pubkey.(*crypto.PubKeyEd25519)
	acc := &Account{
		Pubkey:  pk[:],
		Balance: amount,
		Nonce:   1, // begin from 1
	}
	as.accountCache.Set(pk.KeyString(), acc)
	return nil
}

func (as *AccState) SubBalance(pubkey crypto.PubKey, amount *big.Int) error {
	keystring := pubkey.KeyString()
	as.Lock()
	defer as.Unlock()

	// from cache
	if itfc, ok := as.accountCache.Get(keystring); ok {
		acc := itfc.(*Account)
		if acc.Balance.Cmp(amount) >= 0 {
			acc.Balance = new(big.Int).Sub(acc.Balance, amount)
			return nil
		}
		return errors.New("insufficent balance to sub")
	}

	// from db
	if _, value, exist := as.trie.Get([]byte(keystring)); exist {
		acc := new(Account)
		err := acc.FromBytes(value)
		if err != nil {
			return err
		}
		if acc.Balance.Cmp(amount) >= 0 {
			acc.Balance = new(big.Int).Sub(acc.Balance, amount)
			as.accountCache.Set(keystring, acc)
			return nil
		}
		return errors.New("insufficent balance to sub")
	}

	// Not exist
	return fmt.Errorf("account not exist: %s", keystring)
}

func (as *AccState) IncreaseNonce(pubkey crypto.PubKey, n uint64) error {
	keystring := pubkey.KeyString()
	as.Lock()
	defer as.Unlock()

	// from cache
	if itfc, ok := as.accountCache.Get(keystring); ok {
		acc := itfc.(*Account)
		acc.Nonce += n
		return nil
	}

	// from db
	if _, value, exist := as.trie.Get([]byte(keystring)); exist {
		acc := new(Account)
		err := acc.FromBytes(value)
		if err != nil {
			return err
		}
		acc.Nonce += n
		as.accountCache.Set(keystring, acc)
		return nil
	}

	return fmt.Errorf("account not exist: %s", keystring)
}

func (as *AccState) QueryNonce(pubkey crypto.PubKey) uint64 {
	keystring := pubkey.KeyString()
	as.Lock()
	defer as.Unlock()

	// from cache
	if itfc, ok := as.accountCache.Get(keystring); ok {
		acc := itfc.(*Account)
		return acc.Nonce
	}

	// from db
	if _, value, exist := as.trie.Get([]byte(keystring)); exist {
		acc := new(Account)
		err := acc.FromBytes(value)
		if err != nil {
			log.Println(err)
			return 0
		}
		return acc.Nonce
	}

	return 0
}

func (as *AccState) QueryBalance(pubkey crypto.PubKey) *big.Int {
	keystring := pubkey.KeyString()
	as.Lock()
	defer as.Unlock()

	// from cache
	if itfc, ok := as.accountCache.Get(keystring); ok {
		acc := itfc.(*Account)
		return acc.Balance
	}

	// from db
	if _, value, exist := as.trie.Get([]byte(keystring)); exist {
		acc := new(Account)
		err := acc.FromBytes(value)
		if err != nil {
			log.Println(err)
			return Big0
		}
		return acc.Balance
	}

	return Big0
}

// Commit returns the new root bytes
func (as *AccState) Commit() ([]byte, error) {
	as.mtx.Lock()
	defer as.mtx.Unlock()

	as.accountCache.Exec(func(k string, v interface{}) {
		acc := v.(*Account)
		as.trie.Set([]byte(k), acc.ToBytes())
	})

	as.rootHash = as.trie.Save()
	return as.rootHash, nil
}

// Load dumps all the buffer, start every thing from a clean state
func (as *AccState) Load(root []byte) {
	as.mtx.Lock()
	defer as.mtx.Unlock()
	as.accountCache = mlist.NewMapList()
	as.trie.Load(root)
}

// Reload works the same as Load, just for semantic purpose
func (as *AccState) Reload(root []byte) {
	as.Lock()
	as.accountCache = mlist.NewMapList()
	as.trie.Load(root)
	as.Unlock()
}

func (as *AccState) Iterate(fn func(*Account) bool) {
	as.Lock()
	defer as.Unlock()

	// Iterate cache first
	as.accountCache.Exec(func(key string, value interface{}) {
		acc := value.(*Account)
		fn(acc)
	})

	// Iterate tree
	as.trie.Iterate(func(key, value []byte) bool {
		acc := new(Account)
		if err := acc.FromBytes(value); err != nil {
			fmt.Println("Iterate acc state faild:", err.Error())
			return true
		}

		// escape cache
		var pubkey crypto.PubKeyEd25519
		copy(pubkey[:], acc.Pubkey)
		if _, exist := as.accountCache.Get(pubkey.KeyString()); exist {
			return false
		}

		return fn(acc)
	})
}

func (as *AccState) AssetNonce(addr []byte, n uint64) error {
	acc, err := as.GetAccount(addr)
	if err != nil {
		return err
	}
	if acc.Nonce != n {
		return fmt.Errorf("Invalid nonce or duplicate tx [In-tx %d, In-db %d]", n, acc.Nonce)
	}

	return nil
}

func (as *AccState) Hash() []byte {
	as.Lock()
	defer as.Unlock()

	return as.trie.Hash()
}

func (as *AccState) Size() int {
	as.Lock()
	defer as.Unlock()

	return as.trie.Size()
}

func (oa *Account) FromBytes(bytes []byte) error {
	if err := json.Unmarshal(bytes, oa); err != nil {
		return err
	}
	return nil
}

func (oa *Account) ToBytes() []byte {
	bys, err := json.Marshal(oa)
	if err != nil {
		return nil
	}
	return bys
}
