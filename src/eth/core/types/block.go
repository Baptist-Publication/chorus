// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package types contains data types related to Ethereum consensus.
package types

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/Baptist-Publication/chorus/src/eth/common"
	"github.com/Baptist-Publication/chorus/src/eth/common/hexutil"
	"github.com/Baptist-Publication/chorus/src/eth/crypto/sha3"
	"github.com/Baptist-Publication/chorus/src/eth/rlp"
)

var (
	EmptyRootHash = DeriveSha(Transactions{})
	// EmptyUncleHash = CalcUncleHash(nil)
)

var (
	errMissingHeaderMixDigest = errors.New("missing mixHash in JSON block header")
	errMissingHeaderFields    = errors.New("missing required JSON block header fields")
	errBadNonceSize           = errors.New("invalid block nonce size, want 8 bytes")
)

// A BlockNonce is a 64-bit hash which proves (combined with the
// mix-hash) that a sufficient amount of computation has been carried
// out on a block.
type BlockNonce [8]byte

// EncodeNonce converts the given integer to a block nonce.
func EncodeNonce(i uint64) BlockNonce {
	var n BlockNonce
	binary.BigEndian.PutUint64(n[:], i)
	return n
}

// Uint64 returns the integer value of a block nonce.
func (n BlockNonce) Uint64() uint64 {
	return binary.BigEndian.Uint64(n[:])
}

// MarshalJSON implements json.Marshaler
func (n BlockNonce) MarshalJSON() ([]byte, error) {
	return hexutil.Bytes(n[:]).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler
func (n *BlockNonce) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalJSON("BlockNonce", input, n[:])
}

// Header represents a block header in the Ethereum blockchain.
type Header struct {
	ParentHash  common.Hash    // Hash to the previous block
	UncleHash   common.Hash    // Uncles of this block
	Coinbase    common.Address // The coin base address
	Root        common.Hash    // Block Trie state
	TxHash      common.Hash    // Tx sha
	ReceiptHash common.Hash    // Receipt sha
	Bloom       Bloom          // Bloom
	Difficulty  *big.Int       // Difficulty for the current block
	Number      *big.Int       // The block number
	GasLimit    *big.Int       // Gas limit
	GasUsed     *big.Int       // Gas used
	Time        *big.Int       // Creation time
	Extra       []byte         // Extra data
	MixDigest   common.Hash    // for quick difficulty verification
	Nonce       BlockNonce
}

type jsonHeader struct {
	ParentHash  *common.Hash    `json:"parentHash"`
	UncleHash   *common.Hash    `json:"sha3Uncles"`
	Coinbase    *common.Address `json:"miner"`
	Root        *common.Hash    `json:"stateRoot"`
	TxHash      *common.Hash    `json:"transactionsRoot"`
	ReceiptHash *common.Hash    `json:"receiptsRoot"`
	Bloom       *Bloom          `json:"logsBloom"`
	Difficulty  *hexutil.Big    `json:"difficulty"`
	Number      *hexutil.Big    `json:"number"`
	GasLimit    *hexutil.Big    `json:"gasLimit"`
	GasUsed     *hexutil.Big    `json:"gasUsed"`
	Time        *hexutil.Big    `json:"timestamp"`
	Extra       *hexutil.Bytes  `json:"extraData"`
	MixDigest   *common.Hash    `json:"mixHash"`
	Nonce       *BlockNonce     `json:"nonce"`
}

// Hash returns the block hash of the header, which is simply the keccak256 hash of its
// RLP encoding.
func (h *Header) Hash() common.Hash {
	return rlpHash(h)
}

// HashNoNonce returns the hash which is used as input for the proof-of-work search.
func (h *Header) HashNoNonce() common.Hash {
	return rlpHash([]interface{}{
		h.ParentHash,
		h.UncleHash,
		h.Coinbase,
		h.Root,
		h.TxHash,
		h.ReceiptHash,
		h.Bloom,
		h.Difficulty,
		h.Number,
		h.GasLimit,
		h.GasUsed,
		h.Time,
		h.Extra,
	})
}

// MarshalJSON encodes headers into the web3 RPC response block format.
func (h *Header) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonHeader{
		ParentHash:  &h.ParentHash,
		UncleHash:   &h.UncleHash,
		Coinbase:    &h.Coinbase,
		Root:        &h.Root,
		TxHash:      &h.TxHash,
		ReceiptHash: &h.ReceiptHash,
		Bloom:       &h.Bloom,
		Difficulty:  (*hexutil.Big)(h.Difficulty),
		Number:      (*hexutil.Big)(h.Number),
		GasLimit:    (*hexutil.Big)(h.GasLimit),
		GasUsed:     (*hexutil.Big)(h.GasUsed),
		Time:        (*hexutil.Big)(h.Time),
		Extra:       (*hexutil.Bytes)(&h.Extra),
		MixDigest:   &h.MixDigest,
		Nonce:       &h.Nonce,
	})
}

// UnmarshalJSON decodes headers from the web3 RPC response block format.
func (h *Header) UnmarshalJSON(input []byte) error {
	var dec jsonHeader
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	// Ensure that all fields are set. MixDigest is checked separately because
	// it is a recent addition to the spec (as of August 2016) and older RPC server
	// implementations might not provide it.
	if dec.MixDigest == nil {
		return errMissingHeaderMixDigest
	}
	if dec.ParentHash == nil || dec.UncleHash == nil || dec.Coinbase == nil ||
		dec.Root == nil || dec.TxHash == nil || dec.ReceiptHash == nil ||
		dec.Bloom == nil || dec.Difficulty == nil || dec.Number == nil ||
		dec.GasLimit == nil || dec.GasUsed == nil || dec.Time == nil ||
		dec.Extra == nil || dec.Nonce == nil {
		return errMissingHeaderFields
	}
	// Assign all values.
	h.ParentHash = *dec.ParentHash
	h.UncleHash = *dec.UncleHash
	h.Coinbase = *dec.Coinbase
	h.Root = *dec.Root
	h.TxHash = *dec.TxHash
	h.ReceiptHash = *dec.ReceiptHash
	h.Bloom = *dec.Bloom
	h.Difficulty = (*big.Int)(dec.Difficulty)
	h.Number = (*big.Int)(dec.Number)
	h.GasLimit = (*big.Int)(dec.GasLimit)
	h.GasUsed = (*big.Int)(dec.GasUsed)
	h.Time = (*big.Int)(dec.Time)
	h.Extra = *dec.Extra
	h.MixDigest = *dec.MixDigest
	h.Nonce = *dec.Nonce
	return nil
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
