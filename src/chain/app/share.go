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

type ShareState struct {
	root     []byte
	mtx      sync.RWMutex
	database db.DB
	rootHash []byte
	trie     *merkle.IAVLTree

	//key is ed25519 pubkey
	ShareCache *mlist.MapList
}

type Share struct {
	Pubkey        []byte
	ShareBalance  *big.Int
	ShareGuaranty *big.Int
	MHeight       def.INT
}

func NewShareState(database db.DB) *ShareState {
	return &ShareState{
		//dirty:        make(map[string]struct{}),
		database:   database,
		trie:       merkle.NewIAVLTree(1024, database),
		ShareCache: mlist.NewMapList(),
	}
}

func (ps *ShareState) Copy() *ShareState {
	nps := &ShareState{
		//dirty:        make(map[string]struct{}),
		root:       ps.root,
		database:   ps.database,
		trie:       merkle.NewIAVLTree(1024, ps.database),
		ShareCache: mlist.NewMapList(),
	}
	nps.trie.Load(ps.root)
	return nps
}

func (ps *ShareState) Lock() {
	ps.mtx.Lock()
}

func (ps *ShareState) Unlock() {
	ps.mtx.Unlock()
}

func (ps *ShareState) CreateShareAccount(pubkey []byte, balance *big.Int, height def.INT) {
	pub := crypto.PubKeyEd25519{}
	copy(pub[:], pubkey[:])

	pwr := &Share{
		Pubkey:        pubkey,
		ShareBalance:  new(big.Int).Set(balance),
		ShareGuaranty: big0,
		MHeight:       height,
	}

	ps.ShareCache.Set(pub.KeyString(), pwr)
}

func (ps *ShareState) GetShareAccount(pubkey []byte) *Share {
	pub := crypto.PubKeyEd25519{}
	copy(pub[:], pubkey)

	if pwr, ok := ps.ShareCache.Get(pub.KeyString()); ok {
		return pwr.(*Share)
	}
	if _, Sharebytes, exist := ps.trie.Get([]byte(pub.KeyString())); exist {
		pwr := new(Share)
		pwr.FromBytes(Sharebytes)
		return pwr
	}
	return nil
}

func (ps *ShareState) QueryShare(pubkey crypto.PubKey) (*big.Int, def.INT) {
	keystring := pubkey.KeyString()

	// from cache
	if itfc, ok := ps.ShareCache.Get(keystring); ok {
		pwr := itfc.(*Share)
		return pwr.ShareBalance, pwr.MHeight
	}

	// from db
	if _, value, exist := ps.trie.Get([]byte(keystring)); exist {
		pwr := new(Share)
		err := pwr.FromBytes(value)
		if err != nil {
			log.Println(err)
			return big0, 0
		}
		return pwr.ShareBalance, pwr.MHeight
	}

	return big0, 0
}

func (ps *ShareState) AddShareBalance(pubkey crypto.PubKey, amount *big.Int, height def.INT) error {
	keystring := pubkey.KeyString()

	// from cache
	if itfc, ok := ps.ShareCache.Get(keystring); ok {
		pwr := itfc.(*Share)
		pwr.ShareBalance = new(big.Int).Add(pwr.ShareBalance, amount)
		// pwr.MHeight = height
		return nil
	}

	// from db
	if _, value, exist := ps.trie.Get([]byte(keystring)); exist {
		pwr := new(Share)
		err := pwr.FromBytes(value)
		if err != nil {
			return err
		}
		pwr.ShareBalance = new(big.Int).Add(pwr.ShareBalance, amount)
		// pwr.MHeight = height
		ps.ShareCache.Set(keystring, pwr)
		return nil
	}

	// new account
	pk := pubkey.(*crypto.PubKeyEd25519)
	pwr := &Share{
		Pubkey:        pk[:],
		ShareBalance:  amount,
		ShareGuaranty: big0,
		MHeight:       height,
	}
	ps.ShareCache.Set(pk.KeyString(), pwr)
	return nil
}

func (ps *ShareState) SubShareBalance(pubkey crypto.PubKey, amount *big.Int) error {
	keystring := pubkey.KeyString()

	// from cache
	if itfc, ok := ps.ShareCache.Get(keystring); ok {
		pwr := itfc.(*Share)
		if pwr.ShareBalance.Cmp(amount) >= 0 {
			pwr.ShareBalance = new(big.Int).Sub(pwr.ShareBalance, amount)
			// pwr.MHeight = height
			return nil
		}
		return errors.New("insufficent ShareBalance to sub")
	}

	// from db
	if _, value, exist := ps.trie.Get([]byte(keystring)); exist {
		pwr := new(Share)
		err := pwr.FromBytes(value)
		if err != nil {
			return err
		}
		if pwr.ShareBalance.Cmp(amount) >= 0 {
			pwr.ShareBalance = new(big.Int).Sub(pwr.ShareBalance, amount)
			// pwr.MHeight = height
			ps.ShareCache.Set(keystring, pwr)
			return nil
		}
		return errors.New("insufficent ShareBalance to sub")
	}

	// Not exist
	return fmt.Errorf("Share not exist: %s", keystring)
}

func (ps *ShareState) AddGuaranty(pubkey crypto.PubKey, amount *big.Int, height def.INT) error {
	keystring := pubkey.KeyString()

	// from cache
	if itfc, ok := ps.ShareCache.Get(keystring); ok {
		pwr := itfc.(*Share)
		pwr.ShareGuaranty = new(big.Int).Add(pwr.ShareGuaranty, amount)
		return nil
	}

	// from db
	if _, value, exist := ps.trie.Get([]byte(keystring)); exist {
		pwr := new(Share)
		err := pwr.FromBytes(value)
		if err != nil {
			return err
		}
		pwr.ShareGuaranty = new(big.Int).Add(pwr.ShareGuaranty, amount)
		ps.ShareCache.Set(keystring, pwr)
		return nil
	}

	// new account
	pk := pubkey.(*crypto.PubKeyEd25519)
	pwr := &Share{
		Pubkey:        pk[:],
		ShareBalance:  big0,
		ShareGuaranty: amount,
		MHeight:       height,
	}
	ps.ShareCache.Set(pk.KeyString(), pwr)
	return nil
}

func (ps *ShareState) SubGuaranty(pubkey crypto.PubKey, amount *big.Int) error {
	keystring := pubkey.KeyString()

	// from cache
	if itfc, ok := ps.ShareCache.Get(keystring); ok {
		pwr := itfc.(*Share)
		if pwr.ShareGuaranty.Cmp(amount) >= 0 {
			pwr.ShareGuaranty = new(big.Int).Sub(pwr.ShareGuaranty, amount)
			return nil
		}
		return errors.New("insufficent ShareGuarantee to sub")
	}

	// from db
	if _, value, exist := ps.trie.Get([]byte(keystring)); exist {
		pwr := new(Share)
		err := pwr.FromBytes(value)
		if err != nil {
			return err
		}
		if pwr.ShareGuaranty.Cmp(amount) >= 0 {
			pwr.ShareGuaranty = new(big.Int).Sub(pwr.ShareGuaranty, amount)
			ps.ShareCache.Set(keystring, pwr)
			return nil
		}
		return errors.New("insufficent ShareGuarantee to sub")
	}

	// Not exist
	return fmt.Errorf("Guarantee not exist: %s", keystring)
}

func (ps *ShareState) MarkShare(pubkey crypto.PubKey, mValue def.INT) error {
	keystring := pubkey.KeyString()

	// from cache
	if itfc, ok := ps.ShareCache.Get(keystring); ok {
		pwr := itfc.(*Share)
		pwr.MHeight = mValue
		return nil
	}

	// from db
	if _, value, exist := ps.trie.Get([]byte(keystring)); exist {
		pwr := new(Share)
		err := pwr.FromBytes(value)
		if err != nil {
			return err
		}
		pwr.MHeight = mValue
		ps.ShareCache.Set(keystring, pwr)
		return nil
	}

	return nil
}

// Commit returns the new root bytes
func (ps *ShareState) Commit() ([]byte, error) {
	ps.ShareCache.Exec(func(k string, v interface{}) {
		pwr := v.(*Share)
		if pwr.ShareBalance.Cmp(big0) == 0 && pwr.ShareGuaranty.Cmp(big0) == 0 {
			ps.trie.Remove([]byte(k))
		} else {
			ps.trie.Set([]byte(k), pwr.ToBytes())
		}
	})

	ps.rootHash = ps.trie.Save()
	return ps.rootHash, nil
}

// Load dumps all the buffer, start every thing from a clean state
func (ps *ShareState) Load(root []byte) {
	ps.ShareCache = mlist.NewMapList()
	ps.trie.Load(root)
	ps.root = root
}

// Reload works the same as Load, just for semantic purpose
func (ps *ShareState) Reload(root []byte) {
	ps.ShareCache = mlist.NewMapList()
	ps.trie.Load(root)
	ps.root = root
}

func (ps *ShareState) Iterate(fn func(*Share) bool) {
	// Iterate cache first
	ps.ShareCache.Exec(func(key string, value interface{}) {
		pwr := value.(*Share)
		pub := crypto.PubKeyEd25519{}
		copy(pub[:], pwr.Pubkey[:])
		if pwr.ShareGuaranty.Cmp(big0) != 0 {
			fn(pwr)
		}
	})

	// Iterate tree
	ps.trie.Iterate(func(key, value []byte) bool {
		pwr := new(Share)
		if err := pwr.FromBytes(value); err != nil {
			fmt.Println("Iterate power state faild:", err.Error())
			return true
		}

		// escape cache
		var pubkey crypto.PubKeyEd25519
		copy(pubkey[:], pwr.Pubkey)
		if _, exist := ps.ShareCache.Get(pubkey.KeyString()); exist {
			return false
		}

		if pwr.ShareGuaranty.Cmp(big0) != 0 {
			return fn(pwr)
		}
		return false
	})
}

func (ps *ShareState) Hash() []byte {
	return ps.trie.Hash()
}

func (ps *ShareState) Size() int {
	return ps.trie.Size()
}

func (oa *Share) FromBytes(bytes []byte) error {
	if err := json.Unmarshal(bytes, oa); err != nil {
		return err
	}
	return nil
}

func (oa *Share) ToBytes() []byte {
	bys, err := json.Marshal(oa)
	if err != nil {
		return nil
	}
	return bys
}
