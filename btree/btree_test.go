package btree

import (
    "testing"
	"fmt"
	// "math/rand"
	"btree/node"
)

const items = 1e8

func TestSet(t *testing.T) {
	btree := NewBTree()
	for i:=uint64(0); i<items; i++ {
		btree.Set(i, node.V{Block: uint32(i), Offset: uint32(i/2)})
		if i % (items/20) == 0 {
			fmt.Printf("%v\n", float64(i)/float64(items))
		}
	}
	// fmt.Printf("%v\n", btree.Traverse())
	// btree.Print()
}
