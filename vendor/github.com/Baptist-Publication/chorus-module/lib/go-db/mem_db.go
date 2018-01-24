// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"fmt"
	"sync"
)

func init() {
	registerDBCreator(MemDBBackendStr, func(name string, dir string) (DB, error) {
		return NewMemDB(), nil
	}, false)
}

type MemDB struct {
	mtx sync.Mutex
	db  map[string][]byte
}

func NewMemDB() *MemDB {
	database := &MemDB{db: make(map[string][]byte)}
	return database
}

func (db *MemDB) Get(key []byte) []byte {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	return db.db[string(key)]
}

func (db *MemDB) Set(key []byte, value []byte) {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	db.db[string(key)] = value
}

func (db *MemDB) SetSync(key []byte, value []byte) {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	db.db[string(key)] = value
}

func (db *MemDB) Delete(key []byte) {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	delete(db.db, string(key))
}

func (db *MemDB) DeleteSync(key []byte) {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	delete(db.db, string(key))
}

func (db *MemDB) Close() {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	db = nil
}

func (db *MemDB) Print() {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	for key, value := range db.db {
		fmt.Printf("[%X]:\t[%X]\n", []byte(key), value)
	}
}
func (db *MemDB) Iterator() Iterator {
	//todo
	return nil
}

func (db *MemDB) NewBatch() Batch {
	return &memDBBatch{db, nil}
}

//--------------------------------------------------------------------------------

type memDBBatch struct {
	db  *MemDB
	ops []operation
}

type opType int

const (
	opTypeSet    = 1
	opTypeDelete = 2
)

type operation struct {
	opType
	key   []byte
	value []byte
}

func (mBatch *memDBBatch) Set(key, value []byte) {
	mBatch.ops = append(mBatch.ops, operation{opTypeSet, key, value})
}

func (mBatch *memDBBatch) Delete(key []byte) {
	mBatch.ops = append(mBatch.ops, operation{opTypeDelete, key, nil})
}

func (mBatch *memDBBatch) Write() {
	mBatch.db.mtx.Lock()
	defer mBatch.db.mtx.Unlock()

	for _, op := range mBatch.ops {
		if op.opType == opTypeSet {
			mBatch.db.db[string(op.key)] = op.value
		} else if op.opType == opTypeDelete {
			delete(mBatch.db.db, string(op.key))
		}
	}

}
