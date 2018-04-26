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

package clist

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestSmall(t *testing.T) {
	l := New()
	el1 := l.PushBack(1)
	el2 := l.PushBack(2)
	el3 := l.PushBack(3)
	if l.Len() != 3 {
		t.Error("Expected len 3, got ", l.Len())
	}

	//fmt.Printf("%p %v\n", el1, el1)
	//fmt.Printf("%p %v\n", el2, el2)
	//fmt.Printf("%p %v\n", el3, el3)

	r1 := l.Remove(el1)

	//fmt.Printf("%p %v\n", el1, el1)
	//fmt.Printf("%p %v\n", el2, el2)
	//fmt.Printf("%p %v\n", el3, el3)

	r2 := l.Remove(el2)

	//fmt.Printf("%p %v\n", el1, el1)
	//fmt.Printf("%p %v\n", el2, el2)
	//fmt.Printf("%p %v\n", el3, el3)

	r3 := l.Remove(el3)

	if r1 != 1 {
		t.Error("Expected 1, got ", r1)
	}
	if r2 != 2 {
		t.Error("Expected 2, got ", r2)
	}
	if r3 != 3 {
		t.Error("Expected 3, got ", r3)
	}
	if l.Len() != 0 {
		t.Error("Expected len 0, got ", l.Len())
	}

}

/*
This test is quite hacky because it relies on SetFinalizer
which isn't guaranteed to run at all.
*/
func _TestGCFifo(t *testing.T) {

	const numElements = 1000000
	l := New()
	gcCount := new(uint64)

	// SetFinalizer doesn't work well with circular structures,
	// so we construct a trivial non-circular structure to
	// track.
	type value struct {
		Int int
	}
	done := make(chan struct{})

	for i := 0; i < numElements; i++ {
		v := new(value)
		v.Int = i
		l.PushBack(v)
		runtime.SetFinalizer(v, func(v *value) {
			atomic.AddUint64(gcCount, 1)
		})
	}

	for el := l.Front(); el != nil; {
		l.Remove(el)
		//oldEl := el
		el = el.Next()
		//oldEl.DetachPrev()
		//oldEl.DetachNext()
	}

	runtime.GC()
	time.Sleep(time.Second * 3)
	runtime.GC()
	time.Sleep(time.Second * 3)
	_ = done

	if *gcCount != numElements {
		t.Errorf("Expected gcCount to be %v, got %v", numElements,
			*gcCount)
	}
}

/*
This test is quite hacky because it relies on SetFinalizer
which isn't guaranteed to run at all.
*/
func _TestGCRandom(t *testing.T) {

	const numElements = 1000000
	l := New()
	gcCount := 0

	// SetFinalizer doesn't work well with circular structures,
	// so we construct a trivial non-circular structure to
	// track.
	type value struct {
		Int int
	}

	for i := 0; i < numElements; i++ {
		v := new(value)
		v.Int = i
		l.PushBack(v)
		runtime.SetFinalizer(v, func(v *value) {
			gcCount += 1
		})
	}

	els := make([]*CElement, 0, numElements)
	for el := l.Front(); el != nil; el = el.Next() {
		els = append(els, el)
	}

	for _, i := range rand.Perm(numElements) {
		el := els[i]
		l.Remove(el)
		el = el.Next()
	}

	runtime.GC()
	time.Sleep(time.Second * 3)

	if gcCount != numElements {
		t.Errorf("Expected gcCount to be %v, got %v", numElements,
			gcCount)
	}
}

func TestScanRightDeleteRandom(t *testing.T) {

	const numElements = 10000
	const numTimes = 100000
	const numScanners = 10

	l := New()
	stop := make(chan struct{})

	els := make([]*CElement, numElements, numElements)
	for i := 0; i < numElements; i++ {
		el := l.PushBack(i)
		els[i] = el
	}

	// Launch scanner routines that will rapidly iterate over elements.
	for i := 0; i < numScanners; i++ {
		go func(scannerID int) {
			var el *CElement
			restartCounter := 0
			counter := 0
		FOR_LOOP:
			for {
				select {
				case <-stop:
					fmt.Println("stopped")
					break FOR_LOOP
				default:
				}
				if el == nil {
					el = l.FrontWait()
					restartCounter += 1
				}
				el = el.Next()
				counter += 1
			}
			fmt.Printf("Scanner %v restartCounter: %v counter: %v\n", scannerID, restartCounter, counter)
		}(i)
	}

	// Remove an element, push back an element.
	for i := 0; i < numTimes; i++ {
		// Pick an element to remove
		rmElIdx := rand.Intn(len(els))
		rmEl := els[rmElIdx]

		// Remove it
		l.Remove(rmEl)
		//fmt.Print(".")

		// Insert a new element
		newEl := l.PushBack(-1*i - 1)
		els[rmElIdx] = newEl

		if i%100000 == 0 {
			fmt.Printf("Pushed %vK elements so far...\n", i/1000)
		}

	}

	// Stop scanners
	close(stop)
	time.Sleep(time.Second * 1)

	// And remove all the elements.
	for el := l.Front(); el != nil; el = el.Next() {
		l.Remove(el)
	}
	if l.Len() != 0 {
		t.Fatal("Failed to remove all elements from CList")
	}
}

func TestAppend(t *testing.T) {
	l := New()
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)
	if l.Len() != 3 {
		t.Error("Expected len 3, got ", l.Len())
	}

	l2 := New()
	l2.PushBack(4)
	l2.PushBack(5)
	l2.PushBack(6)
	if l2.Len() != 3 {
		t.Error("Expected len 3, got ", l2.Len())
	}

	l.AppendList(l2)

	fmt.Println("after append")

	for el := l.Front(); el != nil; el = el.Next() {
		fmt.Printf("%p %v\n", el, el.Value.(int))
	}

	l.PushBack(7)
	l.PushBack(8)
	l.PushBack(9)

	fmt.Println("after push")
	for el := l.Front(); el != nil; el = el.Next() {
		fmt.Printf("%p %v\n", el, el.Value.(int))
	}

	fmt.Println("l2:")
	for el := l2.Front(); el != nil; el = el.Next() {
		fmt.Printf("%p %v\n", el, el.Value.(int))
	}

	l2.Clear()

	fmt.Println("l2, after clear:")
	for el := l2.Front(); el != nil; el = el.Next() {
		fmt.Printf("%p %v\n", el, el.Value.(int))
	}

	l2.PushBack(10)
	l2.PushBack(11)
	l2.PushBack(12)

	fmt.Println("l1,after l2 push")
	for el := l.Front(); el != nil; el = el.Next() {
		fmt.Printf("%p %v\n", el, el.Value.(int))
	}

	fmt.Println("l2,after l2 push:")
	for el := l2.Front(); el != nil; el = el.Next() {
		fmt.Printf("%p %v\n", el, el.Value.(int))
	}

	for el := l.Front(); el != nil; el = el.Next() {
		rm := l.Remove(el)
		fmt.Printf("rm : %v\n", rm.(int))
	}

	l.AppendList(l2)
	fmt.Println("l1, l.Removed and append l2:")
	for el := l.Front(); el != nil; el = el.Next() {
		fmt.Printf("%p %v\n", el, el.Value.(int))
	}
	l2.Clear()
}
