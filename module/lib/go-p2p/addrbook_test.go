// Copyright 2017 Baptist-Publication Information Technology Services Co.,Ltd.
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

package p2p

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
)

const addrBookStrict = true

func createTempFileName(prefix string) string {
	f, err := ioutil.TempFile("", prefix)
	if err != nil {
		panic(err)
	}
	fname := f.Name()
	err = f.Close()
	if err != nil {
		panic(err)
	}
	return fname
}

func TestEmpty(t *testing.T) {
	fname := createTempFileName("addrbook_test")
	// t.Logf("New tempfile name: %v", fname)

	// Save an empty book & load it
	book := NewAddrBook(logger, fname, addrBookStrict)
	book.saveToFile(fname)

	book = NewAddrBook(logger, fname, addrBookStrict)
	book.loadFromFile(fname)

	if book.Size() != 0 {
		t.Errorf("Expected 0 addresses, found %v", book.Size())
	}
}

func randIPv4Address() *NetAddress {
	for {
		ip := fmt.Sprintf("%v.%v.%v.%v",
			rand.Intn(254)+1,
			rand.Intn(255),
			rand.Intn(255),
			rand.Intn(255),
		)
		port := rand.Intn(65535-1) + 1
		addr, _ := NewNetAddressString(fmt.Sprintf("%v:%v", ip, port))
		if addr.Routable() {
			return addr
		}
	}
}

func TestSaveAddresses(t *testing.T) {
	fname := createTempFileName("addrbook_test")
	//t.Logf("New tempfile name: %v", fname)

	// Create some random addresses
	randAddrs := []struct {
		addr *NetAddress
		src  *NetAddress
	}{}
	for i := 0; i < 100; i++ {
		addr := randIPv4Address()
		src := randIPv4Address()
		randAddrs = append(randAddrs, struct {
			addr *NetAddress
			src  *NetAddress
		}{
			addr: addr,
			src:  src,
		})
	}

	// Create the book & populate & save
	book := NewAddrBook(logger, fname, addrBookStrict)
	for _, addrSrc := range randAddrs {
		book.AddAddress(addrSrc.addr, addrSrc.src)
	}
	if book.Size() != 100 {
		t.Errorf("Expected 100 addresses, found %v", book.Size())
	}
	book.saveToFile(fname)

	// Reload the book
	book = NewAddrBook(logger, fname, addrBookStrict)
	book.loadFromFile(fname)

	// Test ...

	if book.Size() != 100 {
		t.Errorf("Expected 100 addresses, found %v", book.Size())
	}

	for _, addrSrc := range randAddrs {
		addr := addrSrc.addr
		src := addrSrc.src
		ka := book.addrLookup[addr.String()]
		if ka == nil {
			t.Fatalf("Expected to find KnownAddress %v but wasn't there.", addr)
		}
		if !(ka.Addr.Equals(addr) && ka.Src.Equals(src)) {
			t.Fatalf("KnownAddress doesn't match addr & src")
		}
	}
}

func TestPromoteToOld(t *testing.T) {
	fname := createTempFileName("addrbook_test")
	t.Logf("New tempfile name: %v", fname)

	// Create some random addresses
	randAddrs := []struct {
		addr *NetAddress
		src  *NetAddress
	}{}
	for i := 0; i < 100; i++ {
		addr := randIPv4Address()
		src := randIPv4Address()
		randAddrs = append(randAddrs, struct {
			addr *NetAddress
			src  *NetAddress
		}{
			addr: addr,
			src:  src,
		})
	}

	// Create the book & populate & save
	book := NewAddrBook(logger, fname, addrBookStrict)
	for _, addrSrc := range randAddrs {
		book.AddAddress(addrSrc.addr, addrSrc.src)
	}
	// Attempt all addresses.
	for _, addrSrc := range randAddrs {
		book.MarkAttempt(addrSrc.addr)
	}
	// Promote half of them
	for i, addrSrc := range randAddrs {
		if i%2 == 0 {
			book.MarkGood(addrSrc.addr)
		}
	}
	book.saveToFile(fname)

	// Reload the book
	book = NewAddrBook(logger, fname, addrBookStrict)
	book.loadFromFile(fname)

	// Test ...

	if book.Size() != 100 {
		t.Errorf("Expected 100 addresses, found %v", book.Size())
	}

	// TODO: do more testing :)

	selection := book.GetSelection()
	t.Logf("selection: %v", selection)
}
