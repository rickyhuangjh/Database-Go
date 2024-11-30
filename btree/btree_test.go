package btree

import (
    "testing"
	"fmt"
	// "math/rand"
	"btree/node"
)

const items = 1e7

func TestSet(t *testing.T) {
	cache := node.NewNodeCache()
	btree := NewBTree(cache)
	for i:=uint64(0); i<items; i++ {
		btree.Set(i, i)
		if i % (items/20) == 0 {
			fmt.Printf("%v\n", float64(i)/float64(items))
		}
	}
	// fmt.Printf("%v\n", btree.Traverse())
	// btree.Print()
}

func TestControl(t *testing.T) {
	m := make(map[uint64]uint64)
	for i:=uint64(0); i<items; i++ {
		m[i] = i
		if i % (items/20) == 0 {
			fmt.Printf("%v\n", float64(i)/float64(items))
		}
	}
}
