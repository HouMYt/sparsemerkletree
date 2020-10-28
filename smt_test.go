package SparseMerkleTree

import (
	"fmt"
	"github.com/magiconair/properties/assert"
	"math/big"
	"testing"
)

type hashTest struct {
}

func (sha *hashTest) hash(a []byte) []byte {
	return a
}

func TestSparseMerkleTree_New(t *testing.T) {
	sha := &hashTest{}
	tree := new(SparseMerkleTree).New(3, []byte{123}, true, sha)
	fmt.Println(tree.dummyHashList)
	fmt.Println(tree.leafs)
	fmt.Println(tree.nodes)
	fmt.Println(tree.getPath([]byte{3}))
}

func TestTobitvec(t *testing.T) {
	sha := &hashTest{}
	tree := new(SparseMerkleTree).New(3, []byte{123}, true, sha)
	fmt.Println(tree.toBitVector([]byte{3}))
	n := new(big.Int).SetBytes([]byte{10})
	for i := 0; i < 10; i++ {
		fmt.Println(n.Bit(i))
	}
	fmt.Println(FromBitVector(tree.toBitVector([]byte{3})))
}

func TestSparseMerkleTree_Add(t *testing.T) {
	sha := &hashTest{}
	tree := new(SparseMerkleTree).New(3, []byte{123}, true, sha)
	tree.Add([]byte{3})
	fmt.Println(tree.nodes)
}

func TestSparseMerkleTree_ProveExist(t *testing.T) {
	sha := &hashTest{}
	tree := new(SparseMerkleTree).New(3, []byte{123}, true, sha)
	tree.Add([]byte{3})
	proofExist := tree.ProveExist([]byte{3})
	assert.Equal(t,tree.VerifyExist(proofExist),true)
	proofAbsence := tree.ProveAbsence([]byte{2})
	result := tree.VerifyAbsence(proofAbsence)
	assert.Equal(t,result,true)
}


