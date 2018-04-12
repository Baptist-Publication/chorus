package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus-module/lib/go-db"
	"github.com/Baptist-Publication/chorus-module/lib/go-merkle"
	"github.com/Baptist-Publication/chorus-module/xlib/def"
	"github.com/Baptist-Publication/chorus-module/xlib/mlist"
)

type PowerState struct {
	root     []byte
	mtx      sync.Mutex
	database db.DB
	rootHash []byte
	trie     *merkle.IAVLTree

	//key is ed25519 pubkey
	PowerCache *mlist.MapList
}

type Power struct {
	Pubkey  []byte
	VTPower *big.Int
	MHeight def.INT
}

func NewPowerState(database db.DB) *PowerState {
	return &PowerState{
		//dirty:        make(map[string]struct{}),
		database:   database,
		trie:       merkle.NewIAVLTree(1024, database),
		PowerCache: mlist.NewMapList(),
	}
}

func (ps *PowerState) Copy() *PowerState {
	nps := &PowerState{
		//dirty:        make(map[string]struct{}),
		root:       ps.root,
		database:   ps.database,
		trie:       merkle.NewIAVLTree(1024, ps.database),
		PowerCache: mlist.NewMapList(),
	}
	nps.trie.Load(ps.root)
	return nps
}

func (ps *PowerState) Lock() {
	ps.mtx.Lock()
}

func (ps *PowerState) Unlock() {
	ps.mtx.Unlock()
}

func (ps *PowerState) CreatePower(pubkey []byte, power *big.Int, height def.INT) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	pub := crypto.PubKeyEd25519{}
	copy(pub[:], pubkey[:])

	pwr := &Power{
		Pubkey:  pubkey,
		VTPower: new(big.Int).Set(power),
		MHeight: height,
	}

	ps.PowerCache.Set(pub.KeyString(), pwr)
}

func (ps *PowerState) GetPower(pubkey []byte) (*Power, error) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	pub := crypto.PubKeyEd25519{}
	copy(pub[:], pubkey)

	if pwr, ok := ps.PowerCache.Get(pub.KeyString()); ok {
		return pwr.(*Power), nil
	}
	if _, Powerbytes, exist := ps.trie.Get([]byte(pub.KeyString())); exist {
		pwr := new(Power)
		pwr.FromBytes(Powerbytes)
		return pwr, nil
	}
	return nil, fmt.Errorf("Power not exist: %X", pubkey)
}

func (ps *PowerState) QueryPower(pubkey crypto.PubKey) (*big.Int, def.INT) {
	keystring := pubkey.KeyString()
	ps.Lock()
	defer ps.Unlock()

	// from cache
	if itfc, ok := ps.PowerCache.Get(keystring); ok {
		pwr := itfc.(*Power)
		return pwr.VTPower, pwr.MHeight
	}

	// from db
	if _, value, exist := ps.trie.Get([]byte(keystring)); exist {
		pwr := new(Power)
		err := pwr.FromBytes(value)
		if err != nil {
			log.Println(err)
			return big0, 0
		}
		return pwr.VTPower, pwr.MHeight
	}

	return big0, 0
}

func (ps *PowerState) AddVTPower(pubkey crypto.PubKey, amount *big.Int, height def.INT) error {
	keystring := pubkey.KeyString()
	ps.Lock()
	defer ps.Unlock()

	// from cache
	if itfc, ok := ps.PowerCache.Get(keystring); ok {
		pwr := itfc.(*Power)
		pwr.VTPower = new(big.Int).Add(pwr.VTPower, amount)
		pwr.MHeight = height
		return nil
	}

	// from db
	if _, value, exist := ps.trie.Get([]byte(keystring)); exist {
		pwr := new(Power)
		err := pwr.FromBytes(value)
		if err != nil {
			return err
		}
		pwr.VTPower = new(big.Int).Add(pwr.VTPower, amount)
		pwr.MHeight = height
		ps.PowerCache.Set(keystring, pwr)
		return nil
	}

	// new account
	pk := pubkey.(*crypto.PubKeyEd25519)
	pwr := &Power{
		Pubkey:  pk[:],
		VTPower: amount,
		MHeight: height,
	}
	ps.PowerCache.Set(pk.KeyString(), pwr)
	return nil
}

func (ps *PowerState) SubVTPower(pubkey crypto.PubKey, amount *big.Int, height def.INT) error {
	keystring := pubkey.KeyString()
	ps.Lock()
	defer ps.Unlock()

	// from cache
	if itfc, ok := ps.PowerCache.Get(keystring); ok {
		pwr := itfc.(*Power)
		if pwr.VTPower.Cmp(amount) >= 0 {
			pwr.VTPower = new(big.Int).Sub(pwr.VTPower, amount)
			// pwr.MHeight = height
			return nil
		}
		return errors.New("insufficent VTPower to sub")
	}

	// from db
	if _, value, exist := ps.trie.Get([]byte(keystring)); exist {
		pwr := new(Power)
		err := pwr.FromBytes(value)
		if err != nil {
			return err
		}
		if pwr.VTPower.Cmp(amount) >= 0 {
			pwr.VTPower = new(big.Int).Sub(pwr.VTPower, amount)
			// pwr.MHeight = height
			ps.PowerCache.Set(keystring, pwr)
			return nil
		}
		return errors.New("insufficent VTPower to sub")
	}

	// Not exist
	return fmt.Errorf("Power not exist: %s", keystring)
}

func (ps *PowerState) MarkPower(pubkey crypto.PubKey, mValue def.INT) error {
	keystring := pubkey.KeyString()
	ps.Lock()
	defer ps.Unlock()

	// from cache
	if itfc, ok := ps.PowerCache.Get(keystring); ok {
		pwr := itfc.(*Power)
		pwr.MHeight = mValue
		return nil
	}

	// from db
	if _, value, exist := ps.trie.Get([]byte(keystring)); exist {
		pwr := new(Power)
		err := pwr.FromBytes(value)
		if err != nil {
			return err
		}
		pwr.MHeight = mValue
		ps.PowerCache.Set(keystring, pwr)
		return nil
	}

	return nil
}

// Commit returns the new root bytes
func (ps *PowerState) Commit() ([]byte, error) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	ps.PowerCache.Exec(func(k string, v interface{}) {
		pwr := v.(*Power)
		if pwr.VTPower.Cmp(big0) == 0 {
			ps.trie.Remove([]byte(k))
		} else {
			ps.trie.Set([]byte(k), pwr.ToBytes())
		}
	})

	ps.rootHash = ps.trie.Save()
	return ps.rootHash, nil
}

// Load dumps all the buffer, start every thing from a clean state
func (ps *PowerState) Load(root []byte) {
	ps.Lock()
	ps.PowerCache = mlist.NewMapList()
	ps.trie.Load(root)
	ps.root = root
	ps.Unlock()
}

// Reload works the same as Load, just for semantic purpose
func (ps *PowerState) Reload(root []byte) {
	ps.Lock()
	ps.PowerCache = mlist.NewMapList()
	ps.trie.Load(root)
	ps.root = root
	ps.Unlock()
}

func (ps *PowerState) Iterate(fn func(*Power) bool) {
	ps.Lock()
	defer ps.Unlock()

	// Iterate cache first
	ps.PowerCache.Exec(func(key string, value interface{}) {
		pwr := value.(*Power)
		if pwr.VTPower.Cmp(big0) != 0 {
			fn(pwr)
		}
	})

	// Iterate tree
	ps.trie.Iterate(func(key, value []byte) bool {
		pwr := new(Power)
		if err := pwr.FromBytes(value); err != nil {
			fmt.Println("Iterate power state faild:", err.Error())
			return true
		}

		// escape cache
		var pubkey crypto.PubKeyEd25519
		copy(pubkey[:], pwr.Pubkey)
		if _, exist := ps.PowerCache.Get(pubkey.KeyString()); exist {
			return false
		}

		return fn(pwr)
	})
}

func (ps *PowerState) Hash() []byte {
	ps.Lock()
	defer ps.Unlock()

	return ps.trie.Hash()
}

func (ps *PowerState) Size() int {
	ps.Lock()
	defer ps.Unlock()

	return ps.trie.Size()
}

func (oa *Power) FromBytes(bytes []byte) error {
	if err := json.Unmarshal(bytes, oa); err != nil {
		return err
	}
	return nil
}

func (oa *Power) ToBytes() []byte {
	bys, err := json.Marshal(oa)
	if err != nil {
		return nil
	}
	return bys
}
