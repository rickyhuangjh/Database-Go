package btree

import (
    "testing"
	"fmt"
	// "math/rand"
	"btree/node"
)

const items = 1e8

func TestSet(t *testing.T) {
	cache := node.NewNodeCacheImpl()
	btree := NewBTree(cache)
	for i:=uint64(0); i<items; i++ {
		btree.Set(1e8-i, i)
		if i % (items/20) == 0 {
			fmt.Printf("%v\n", float64(i)/float64(items))
			fmt.Println(len(cache.Nodes))
		}
	}
	// fmt.Printf("%v\n", btree.Traverse())
	// btree.Print()
}

func TestControl(t *testing.T) {
	m := make(map[uint64]uint64)
	for i:=uint64(0); i<items; i++ {
		m[i] = i
	}
}
