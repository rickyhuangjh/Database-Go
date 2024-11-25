package btree

import (
	"fmt"
	"btree/node"
)


type BTree interface {
	Set(key uint64, val node.V) error
	Delete(key uint64) error
	Get(key uint64) (node.V, bool)
	GetRange(start, end uint64) []node.V
	Traverse() []node.V
	Print()
	Verify()
}

type btree struct {
	Order uint64
	Root  node.BTreeNode
	Size  int
}


func NewBTree() BTree {
	return &btree{
		Root:  nil,
		Size:  0,
	}
}

func (b *btree) Get(key uint64) (node.V, bool) {
	return b.Root.Get(key)
}

func (b *btree) Traverse() []node.V {
	if b.Root == nil {
		return nil
	}
	res := b.Root.Traverse(make([]node.V, 0, b.Size))
	return res
}

func (b *btree) Set(key uint64, val node.V) error {
	if b.Root == nil {
		b.Root = node.NewLeafNode()
	}

	wasAdded, err := b.Root.Set(key, val)
	if b.Root.GetParent() != nil {
		b.Root = b.Root.GetParent()
	}
	if wasAdded {
		b.Size++
	}
	return err
}

func (b *btree) Delete(key uint64) error {
	if b.Root == nil {
		return nil
	}
	wasDeleted, err := b.Root.Delete(key)
	if wasDeleted {
		b.Size--
	}
	b.Root = b.Root.GetNewRoot()
	if b.Root != nil {
		b.Root.SetParent(nil)
	}
	return err
}

func (b *btree) Print() {
	if b.Root == nil {
		fmt.Println("Empty tree")
		return
	}
	b.Root.Print(0)
}

func (b *btree) Verify() {
	if b.Root == nil {
		return
	}
	b.Root.Verify()
}

func (b *btree) GetRange(start, end uint64) []node.V {
	res := make([]node.V, 0)
	res = b.Root.GetRange(start, end, res)
	return res
}
