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

package merkle

import (
	"bytes"

	"golang.org/x/crypto/ripemd160"

	. "github.com/Baptist-Publication/chorus/module/lib/go-common"
	"github.com/Baptist-Publication/chorus/module/lib/go-wire"
)

const proofLimit = 1 << 16 // 64 KB

type IAVLProof struct {
	LeafNode   IAVLProofLeafNode
	InnerNodes []IAVLProofInnerNode
	RootHash   []byte
}

func (proof *IAVLProof) Key() []byte {
	return proof.LeafNode.KeyBytes
}

func (proof *IAVLProof) Value() []byte {
	return proof.LeafNode.ValueBytes
}

func (proof *IAVLProof) Root() []byte {
	return proof.RootHash
}

func (proof *IAVLProof) Valid() bool {
	hash := proof.LeafNode.Hash()
	for _, branch := range proof.InnerNodes {
		hash = branch.Hash(hash)
	}
	return bytes.Equal(proof.RootHash, hash)
}

// LoadProof will deserialize a IAVLProof from bytes
func LoadProof(data []byte) (*IAVLProof, error) {
	// TODO: make go-wire never panic
	n, err := int(0), error(nil)
	proof := wire.ReadBinary(&IAVLProof{}, bytes.NewBuffer(data), proofLimit, &n, &err).(*IAVLProof)
	return proof, err
}

type IAVLProofInnerNode struct {
	Height int8
	Size   int
	Left   []byte
	Right  []byte
}

func (branch IAVLProofInnerNode) Hash(childHash []byte) []byte {
	hasher := ripemd160.New()
	buf := new(bytes.Buffer)
	n, err := int(0), error(nil)
	wire.WriteInt8(branch.Height, buf, &n, &err)
	wire.WriteVarint(branch.Size, buf, &n, &err)
	if len(branch.Left) == 0 {
		wire.WriteByteSlice(childHash, buf, &n, &err)
		wire.WriteByteSlice(branch.Right, buf, &n, &err)
	} else {
		wire.WriteByteSlice(branch.Left, buf, &n, &err)
		wire.WriteByteSlice(childHash, buf, &n, &err)
	}
	if err != nil {
		PanicCrisis(Fmt("Failed to hash IAVLProofInnerNode: %v", err))
	}
	// fmt.Printf("InnerNode hash bytes: %X\n", buf.Bytes())
	hasher.Write(buf.Bytes())
	return hasher.Sum(nil)
}

type IAVLProofLeafNode struct {
	KeyBytes   []byte
	ValueBytes []byte
}

func (leaf IAVLProofLeafNode) Hash() []byte {
	hasher := ripemd160.New()
	buf := new(bytes.Buffer)
	n, err := int(0), error(nil)
	wire.WriteInt8(0, buf, &n, &err)
	wire.WriteVarint(1, buf, &n, &err)
	wire.WriteByteSlice(leaf.KeyBytes, buf, &n, &err)
	wire.WriteByteSlice(leaf.ValueBytes, buf, &n, &err)
	if err != nil {
		PanicCrisis(Fmt("Failed to hash IAVLProofLeafNode: %v", err))
	}
	// fmt.Printf("LeafNode hash bytes:   %X\n", buf.Bytes())
	hasher.Write(buf.Bytes())
	return hasher.Sum(nil)
}

func (node *IAVLNode) constructProof(t *IAVLTree, key []byte, proof *IAVLProof) (exists bool) {
	if node.height == 0 {
		if bytes.Compare(node.key, key) == 0 {
			leaf := IAVLProofLeafNode{
				KeyBytes:   node.key,
				ValueBytes: node.value,
			}
			proof.LeafNode = leaf
			return true
		} else {
			return false
		}
	} else {
		if bytes.Compare(key, node.key) < 0 {
			exists := node.getLeftNode(t).constructProof(t, key, proof)
			if !exists {
				return false
			}
			branch := IAVLProofInnerNode{
				Height: node.height,
				Size:   node.size,
				Left:   nil,
				Right:  node.getRightNode(t).hash,
			}
			proof.InnerNodes = append(proof.InnerNodes, branch)
			return true
		} else {
			exists := node.getRightNode(t).constructProof(t, key, proof)
			if !exists {
				return false
			}
			branch := IAVLProofInnerNode{
				Height: node.height,
				Size:   node.size,
				Left:   node.getLeftNode(t).hash,
				Right:  nil,
			}
			proof.InnerNodes = append(proof.InnerNodes, branch)
			return true
		}
	}
}

// Returns nil if key is not in tree.
func (t *IAVLTree) ConstructProof(key []byte) *IAVLProof {
	if t.root == nil {
		return nil
	}
	t.root.hashWithCount(t) // Ensure that all hashes are calculated.
	proof := &IAVLProof{
		RootHash: t.root.hash,
	}
	exists := t.root.constructProof(t, key, proof)
	if exists {
		return proof
	} else {
		return nil
	}
}
