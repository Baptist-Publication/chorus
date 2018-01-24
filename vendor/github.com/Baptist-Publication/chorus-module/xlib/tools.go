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

package xlib

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"sort"
	"unsafe"
)

func CheckItfcNil(itfc interface{}) bool {
	d := (*struct {
		itab uintptr
		data uintptr
	})(unsafe.Pointer(&itfc))
	return d.data == 0
}

func BinRead(reader io.Reader, data interface{}) error {
	return binary.Read(reader, binary.LittleEndian, data)
}

func BinWrite(writer io.Writer, data interface{}) error {
	return binary.Write(writer, binary.LittleEndian, data)
}

func ReadBytes(reader io.Reader) ([]byte, error) {
	byLen, err := IntFrom4Bytes(reader)
	if err != nil {
		return nil, err
	}
	bys := make([]byte, byLen)
	err = BinRead(reader, &bys)
	return bys, err
}

func WriteBytes(writer io.Writer, bys []byte) error {
	err := BinWrite(writer, uint32(len(bys)))
	if err != nil {
		return err
	}
	return BinWrite(writer, bys)
}

func IntFrom4Bytes(reader io.Reader) (int, error) {
	var num uint32
	if err := BinRead(reader, &num); err != nil {
		return 0, err
	}
	return int(num), nil
}

func IntTo4Bytes(num int) ([]byte, error) {
	var buf bytes.Buffer
	if err := BinWrite(&buf, uint32(num)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src) //
}

type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p Int64Slice) Sort() { sort.Sort(p) }
