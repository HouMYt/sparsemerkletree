package SparseMerkleTree

import (
	"bytes"
	"crypto/sha256"
	"math/big"
)

type SparseMerkleTree struct {
	salted        bool
	hash          hash
	height        uint32
	nodes         [][]byte
	leafs         [][]byte
	dummy         []byte
	dummyHashList [][]byte
}

type ExistProof struct {
	path  [][]byte
	index int
	data  []byte
	root  []byte
}

type hash interface {
	hash([]byte) []byte
}
type hash256 struct {
}

func (sha *hash256) hash(a []byte) []byte {
	hash := sha256.Sum256(a)
	return hash[:]
}

func (t *SparseMerkleTree) New(height uint32, dummy []byte, salt bool,hash hash) *SparseMerkleTree {
	tree := &SparseMerkleTree{
		salted:        salt,
		hash:          hash,
		height:        height,
		nodes:         nil,
		leafs:         nil,
		dummy:         dummy,
		dummyHashList: nil,
	}
	tree.dummyHashList = make([][]byte, height)
	dummyHash := hash.hash(dummy)
	for i := range tree.dummyHashList {
		tree.dummyHashList[i] = dummyHash
		left := make([]byte,len(dummyHash))
		right := make([]byte,len(dummyHash))
		copy(left[:],dummyHash)
		copy(right[:],dummyHash)
		if tree.salted {
			left = append(left, 0)
			right = append(right, 1)
		}
		dummyHash = hash.hash(append(left, right...))
	}
	branchLen := 1 << (tree.height - 1)
	tree.nodes = make([][]byte, 2*branchLen-1)
	h := uint32(0)
	thred := 0
	for i := 0; i < 2*branchLen-1; i++ {
		if i >= thred+(1<<h) {
			thred = i
			h++
		}
		tree.nodes[i] = make([]byte,len(tree.dummyHashList[tree.height-1-h]))
		copy(tree.nodes[i],tree.dummyHashList[tree.height-1-h])
		if tree.salted {
			if i&1 == 1 {
				tree.nodes[i] = append(tree.nodes[i], 0)
			} else {
				tree.nodes[i] = append(tree.nodes[i], 1)
			}
		}
	}

	tree.leafs = make([][]byte, branchLen)
	for i := range tree.leafs {
		tree.leafs[i] = dummy
	}

	return tree
}

func (t *SparseMerkleTree) Add(data []byte) {
	dataHash := t.hash.hash(data)
	path := t.getPath(data)
	if t.salted {
		if path[0]&1 == 1 {
			dataHash = append(dataHash, 0)
		} else {
			dataHash = append(dataHash, 1)
		}
	}
	t.nodes[path[0]] = dataHash
	for i := 1; i < int(t.height)-1; i++ {
		if path[i-1]&1 == 1 {
			left := t.nodes[path[i-1]]
			right := t.nodes[path[i-1]+1]
			dataHash = t.hash.hash(append(left, right...))
		} else {
			right := t.nodes[path[i-1]]
			left := t.nodes[path[i-1]-1]
			dataHash = t.hash.hash(append(left, right...))
		}
		if t.salted {
			if path[i]&1 == 1 {
				dataHash = append(dataHash, 0)
			} else {
				dataHash = append(dataHash, 1)
			}
		}
		t.nodes[path[i]] = dataHash
	}
	t.nodes[0] = t.hash.hash(append(t.nodes[1], t.nodes[2]...))
	t.nodes[0] = append(t.nodes[0], 1)
}

func (t *SparseMerkleTree) ProveExist(n []byte) ExistProof {
	dataHash := t.hash.hash(n)
	path := t.getPath(dataHash)
	proof := ExistProof{}
	proof.path = make([][]byte, 0)
	proof.index = path[0]
	for i := range path {
		proof.path = append(proof.path, t.getSibling(path[i]))
	}
	proof.data = n
	proof.root = t.Root()
	return proof
}

func (t *SparseMerkleTree) ProveAbsence(n []byte) ExistProof {
	proof := t.ProveExist(n)
	return proof
}

func (t *SparseMerkleTree) VerifyExist( proof ExistProof) bool {
	if bytes.Compare(proof.root, t.nodes[0]) != 0 {
		return false
	}
	dataHash := t.hash.hash(proof.data)
	path := t.getPath(proof.data)
	var left []byte
	var right []byte
	for i := range proof.path {
		if t.salted {
			if path[i]&1 == 1 {
				dataHash = append(dataHash, 0)
			} else {
				dataHash = append(dataHash, 1)
			}
		}
		if path[i]&1 == 1 {
			right = proof.path[i]
			left = dataHash
		} else {
			right = dataHash
			left = proof.path[i]
		}
		dataHash = t.hash.hash(append(left, right...))
	}
	dataHash = append(dataHash,1)
	if bytes.Compare(dataHash, proof.root) != 0 {
		return false
	}

	return true
}
func (t *SparseMerkleTree)VerifyAbsence(proof ExistProof)bool  {
	if bytes.Compare(proof.root, t.nodes[0]) != 0 {
		return false
	}
	dataHash := t.hash.hash(t.dummy)
	path := t.getPath(proof.data)
	var left []byte
	var right []byte
	for i := range proof.path {
		if t.salted {
			if path[i]&1 == 1 {
				dataHash = append(dataHash, 0)
			} else {
				dataHash = append(dataHash, 1)
			}
		}
		if path[i]&1 == 1 {
			right = proof.path[i]
			left = dataHash
		} else {
			right = dataHash
			left = proof.path[i]
		}
		dataHash = t.hash.hash(append(left, right...))
	}
	dataHash = append(dataHash,1)
	if bytes.Compare(dataHash, proof.root) != 0 {
		return false
	}

	return true
}

//func (t *SparseMerkleTree) VerifyAbsence( proof ) bool {
//	left,right := t.getSurround(proof.data)
//	if bytes.Compare(left, proof.proofLeft.data) != 0 ||
//		bytes.Compare(right, proof.proofRight.data) != 0 {
//		return false
//	}
//	if !t.VerifyExist(proof.proofLeft) {
//		return false
//	}
//	if !t.VerifyExist(proof.proofRight) {
//		return false
//	}
//	return true
//}

func (t *SparseMerkleTree) Root() []byte {
	return t.nodes[0]
}

func (t *SparseMerkleTree) toBitVector(data []byte) []uint {
	dataInt := new(big.Int).SetBytes(data)
	bitVector := make([]uint, 0)
	for i := 0; i < int(t.height)-1; i++ {
		bitVector = append(bitVector, dataInt.Bit(i))
	}
	return bitVector
}

func (t *SparseMerkleTree) getPath(data []byte) []int {
	bitVector := t.toBitVector(data)
	path := make([]int, t.height-1)
	for h := 0; h < int(t.height)-1; h++ {
		index := (1 << (int(t.height) - h - 1)) - 1
		k := FromBitVector(bitVector[h:])
		//if h > 0 && bitVector[h-1] == 1 {
		//	k = k + 1
		//}
		index = index + int(k)
		path[h] = index
	}
	return path
}

func FromBitVector(a []uint) uint64 {
	result := uint64(0)
	for i := range a {
		if a[i] == 1 {
			result = result + (1 << i)
		}
	}
	return result
}

func (t *SparseMerkleTree) getSibling(index int) []byte {
	if index&1 == 1 {
		return t.nodes[index+1]
	} else {
		return t.nodes[index-1]
	}
}

func (t *SparseMerkleTree) getSurround(data []byte) ([]byte, []byte) {
	dataInt := new(big.Int).SetBytes(data)
	left := new(big.Int).Add(dataInt, new(big.Int).SetInt64(int64(1)))
	right := new(big.Int).Sub(dataInt, new(big.Int).SetInt64(int64(1)))
	return left.Bytes(), right.Bytes()
}
