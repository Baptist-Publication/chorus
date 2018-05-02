package app

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/Baptist-Publication/chorus/eth/rlp"
	"github.com/Baptist-Publication/chorus/module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus/module/lib/go-db"
	"github.com/Baptist-Publication/chorus/module/lib/go-merkle"
	"github.com/Baptist-Publication/chorus/module/xlib/mlist"
)

type Share struct {
	Pubkey        []byte
	ShareBalance  *big.Int
	ShareGuaranty *big.Int
	GHeight       uint64
}

func (shr *Share) Copy() *Share {
	return &Share{
		Pubkey:        shr.Pubkey,
		ShareBalance:  shr.ShareBalance,
		ShareGuaranty: shr.ShareGuaranty,
		GHeight:       shr.GHeight,
	}
}

func (shr *Share) Less(o interface{}) bool {
	return shr.ShareGuaranty.Cmp(o.(*Share).ShareGuaranty) < 0
}

func (shr *Share) FromBytes(bs []byte) error {
	err := rlp.DecodeBytes(bs, shr)
	// err := json.Unmarshal(bs, shr)
	if err != nil {
		return err
	}
	return nil
}

func (shr *Share) ToBytes() []byte {
	bs, err := rlp.EncodeToBytes(shr)
	// bs, err := json.Marshal(shr)
	if err != nil {
		return nil
	}
	return bs
}

type ShareState struct {
	root     []byte
	mtx      sync.RWMutex
	database db.DB
	rootHash []byte
	trie     *merkle.IAVLTree

	ShareCache *mlist.MapList
}

func NewShareState(database db.DB) *ShareState {
	return &ShareState{
		//dirty:        make(map[string]struct{}),
		database:   database,
		trie:       merkle.NewIAVLTree(1024, database),
		ShareCache: mlist.NewMapList(),
	}
}

func (ss *ShareState) Copy() *ShareState {
	nss := &ShareState{
		//dirty:        make(map[string]struct{}),
		root:       ss.root,
		database:   ss.database,
		trie:       merkle.NewIAVLTree(1024, ss.database),
		ShareCache: mlist.NewMapList(),
	}
	nss.trie.Load(ss.root)
	return nss
}

func (ss *ShareState) Lock() {
	ss.mtx.Lock()
}

func (ss *ShareState) Unlock() {
	ss.mtx.Unlock()
}

func (ss *ShareState) CreateShareAccount(pubkey []byte, balance *big.Int) {
	pub := crypto.PubKeyEd25519{}
	copy(pub[:], pubkey[:])
	if shr, ok := ss.ShareCache.Get(pub.KeyString()); ok {
		value := shr.(*Share)
		value.ShareBalance = new(big.Int).Set(balance)
		ss.ShareCache.Set(pub.KeyString(), value)
		return
	}
	shr := &Share{
		Pubkey:        pubkey,
		ShareBalance:  new(big.Int).Set(balance),
		ShareGuaranty: big0,
	}

	ss.ShareCache.Set(pub.KeyString(), shr)
}

func (ss *ShareState) GetShareAccount(pubkey []byte) *Share {
	pub := crypto.PubKeyEd25519{}
	copy(pub[:], pubkey)

	if shr, ok := ss.ShareCache.Get(pub.KeyString()); ok {
		return shr.(*Share)
	}
	if _, Sharebytes, exist := ss.trie.Get([]byte(pub.KeyString())); exist {
		shr := new(Share)
		shr.FromBytes(Sharebytes)
		return shr
	}
	return nil
}

func (ss *ShareState) QueryShare(pubkey crypto.PubKey) (*big.Int, uint64) {
	keystring := pubkey.KeyString()

	// from cache
	if itfc, ok := ss.ShareCache.Get(keystring); ok {
		shr := itfc.(*Share)
		return shr.ShareBalance, shr.GHeight
	}

	// from db
	if _, value, exist := ss.trie.Get([]byte(keystring)); exist {
		shr := new(Share)
		err := shr.FromBytes(value)
		if err != nil {
			log.Println(err)
			return big0, 0
		}
		return shr.ShareBalance, shr.GHeight
	}

	return big0, 0
}

func (ss *ShareState) AddShareBalance(pubkey crypto.PubKey, amount *big.Int) error {
	keystring := pubkey.KeyString()

	// from cache
	if itfc, ok := ss.ShareCache.Get(keystring); ok {
		shr := itfc.(*Share)
		shr.ShareBalance = new(big.Int).Add(shr.ShareBalance, amount)
		return nil
	}

	// from db
	if _, value, exist := ss.trie.Get([]byte(keystring)); exist {
		shr := new(Share)
		err := shr.FromBytes(value)
		if err != nil {
			return err
		}
		shr.ShareBalance = new(big.Int).Add(shr.ShareBalance, amount)
		ss.ShareCache.Set(keystring, shr)
		return nil
	}

	// new account
	pk := pubkey.(*crypto.PubKeyEd25519)
	shr := &Share{
		Pubkey:        pk[:],
		ShareBalance:  amount,
		ShareGuaranty: big0,
	}
	ss.ShareCache.Set(pk.KeyString(), shr)
	return nil
}

func (ss *ShareState) SubShareBalance(pubkey crypto.PubKey, amount *big.Int) error {
	keystring := pubkey.KeyString()

	// from cache
	if itfc, ok := ss.ShareCache.Get(keystring); ok {
		shr := itfc.(*Share)
		if shr.ShareBalance.Cmp(amount) >= 0 {
			shr.ShareBalance = new(big.Int).Sub(shr.ShareBalance, amount)
			return nil
		}
		return errors.New("insufficent ShareBalance to sub")
	}

	// from db
	if _, value, exist := ss.trie.Get([]byte(keystring)); exist {
		shr := new(Share)
		err := shr.FromBytes(value)
		if err != nil {
			return err
		}
		if shr.ShareBalance.Cmp(amount) >= 0 {
			shr.ShareBalance = new(big.Int).Sub(shr.ShareBalance, amount)
			ss.ShareCache.Set(keystring, shr)
			return nil
		}
		return errors.New("insufficent ShareBalance to sub")
	}

	// Not exist
	return fmt.Errorf("Share not exist: %s", keystring)
}

func (ss *ShareState) AddGuaranty(pubkey crypto.PubKey, amount *big.Int, height uint64) error {
	keystring := pubkey.KeyString()

	// from cache
	if itfc, ok := ss.ShareCache.Get(keystring); ok {
		shr := itfc.(*Share)
		shr.ShareGuaranty = new(big.Int).Add(shr.ShareGuaranty, amount)
		shr.GHeight = height
		return nil
	}

	// from db
	if _, value, exist := ss.trie.Get([]byte(keystring)); exist {
		shr := new(Share)
		err := shr.FromBytes(value)
		if err != nil {
			return err
		}
		shr.ShareGuaranty = new(big.Int).Add(shr.ShareGuaranty, amount)
		shr.GHeight = height
		ss.ShareCache.Set(keystring, shr)
		return nil
	}

	// new account
	pk := pubkey.(*crypto.PubKeyEd25519)
	shr := &Share{
		Pubkey:        pk[:],
		ShareBalance:  big0,
		ShareGuaranty: amount,
		GHeight:       height,
	}
	ss.ShareCache.Set(pk.KeyString(), shr)
	return nil
}

func (ss *ShareState) SubGuaranty(pubkey crypto.PubKey, amount *big.Int) error {
	keystring := pubkey.KeyString()

	// from cache
	if itfc, ok := ss.ShareCache.Get(keystring); ok {
		shr := itfc.(*Share)
		if shr.ShareGuaranty.Cmp(amount) >= 0 {
			shr.ShareGuaranty = new(big.Int).Sub(shr.ShareGuaranty, amount)
			return nil
		}
		return errors.New("insufficent ShareGuarantee to sub")
	}

	// from db
	if _, value, exist := ss.trie.Get([]byte(keystring)); exist {
		shr := new(Share)
		err := shr.FromBytes(value)
		if err != nil {
			return err
		}
		if shr.ShareGuaranty.Cmp(amount) >= 0 {
			shr.ShareGuaranty = new(big.Int).Sub(shr.ShareGuaranty, amount)
			ss.ShareCache.Set(keystring, shr)
			return nil
		}
		return errors.New("insufficent ShareGuarantee to sub")
	}

	// Not exist
	return fmt.Errorf("Guarantee not exist: %s", keystring)
}

func (ss *ShareState) MarkShare(pubkey crypto.PubKey, gValue uint64) error {
	keystring := pubkey.KeyString()

	// from cache
	if itfc, ok := ss.ShareCache.Get(keystring); ok {
		shr := itfc.(*Share)
		shr.GHeight = gValue
		return nil
	}

	// from db
	if _, value, exist := ss.trie.Get([]byte(keystring)); exist {
		shr := new(Share)
		err := shr.FromBytes(value)
		if err != nil {
			return err
		}
		shr.GHeight = gValue
		ss.ShareCache.Set(keystring, shr)
		return nil
	}

	return nil
}

// Commit returns the new root bytes
func (ss *ShareState) Commit() ([]byte, error) {
	ss.ShareCache.Exec(func(k string, v interface{}) {
		shr := v.(*Share)
		if shr.ShareBalance.Cmp(big0) == 0 && shr.ShareGuaranty.Cmp(big0) == 0 {
			ss.trie.Remove([]byte(k))
		} else {
			ss.trie.Set([]byte(k), shr.ToBytes())
		}
	})

	ss.rootHash = ss.trie.Save()
	return ss.rootHash, nil
}

// Load dumss all the buffer, start every thing from a clean state
func (ss *ShareState) Load(root []byte) {
	ss.ShareCache = mlist.NewMapList()
	ss.trie.Load(root)
	ss.root = root
}

// Reload works the same as Load, just for semantic purpose
func (ss *ShareState) Reload(root []byte) {
	ss.ShareCache = mlist.NewMapList()
	ss.trie.Load(root)
	ss.root = root
}

func (ss *ShareState) Iterate(fn func(*Share) bool) {
	// Iterate cache first
	ss.ShareCache.Exec(func(key string, value interface{}) {
		shr := value.(*Share)
		pub := crypto.PubKeyEd25519{}
		copy(pub[:], shr.Pubkey[:])
		if shr.ShareGuaranty.Cmp(big0) != 0 {
			fn(shr)
		}
	})

	// Iterate tree
	ss.trie.Iterate(func(key, value []byte) bool {
		shr := new(Share)
		if err := shr.FromBytes(value); err != nil {
			fmt.Println("Iterate share state faild:", err.Error())
			return true
		}

		// escape cache
		var pubkey crypto.PubKeyEd25519
		copy(pubkey[:], shr.Pubkey)
		if _, exist := ss.ShareCache.Get(pubkey.KeyString()); exist {
			return false
		}

		if shr.ShareGuaranty.Cmp(big0) != 0 {
			return fn(shr)
		}
		return false
	})
}

func (ss *ShareState) Hash() []byte {
	return ss.trie.Hash()
}

func (ss *ShareState) Size() int {
	return ss.trie.Size()
}
