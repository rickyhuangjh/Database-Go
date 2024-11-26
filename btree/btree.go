package btree

import (
	"fmt"
	"btree/node"
)


type BTree interface {
	Set(key uint64, val uint64) error
	Delete(key uint64) error
	Get(key uint64) (uint64, bool)
	GetRange(start, end uint64) []uint64
	Traverse() []uint64
	Print()
}

type btree struct {
	Cache node.NodeCache
	RootID uint64
	Size  int
}


func NewBTree(cache node.NodeCache) BTree {
	return &btree{
		Cache: cache,
		RootID:  0,
		Size:  0,
	}
}

func (b *btree) Get(key uint64) (uint64, bool) {
	if b.RootID == 0 {
		return 0, false
	}
	return b.Cache.Get(b.RootID).Get(b.Cache, key)
}

func (b *btree) Traverse() []uint64 {
	if b.RootID == 0 {
		return nil
	}
	res := b.Cache.Get(b.RootID).Traverse(b.Cache, make([]uint64, 0, b.Size))
	return res
}

func (b *btree) Set(key uint64, val uint64) error {
	var root node.BTreeNode
	if b.RootID == 0 {
		root = node.NewLeafNode()
		b.Cache.Register(root)
		b.RootID = root.(*node.LeafNode).ID
	} else {
		root = b.Cache.Get(b.RootID)
	}

	added, err := root.Set(b.Cache, key, val)
	if root.GetParentID() != 0 {
		b.RootID = root.GetParentID()
	}
	if added {
		b.Size++
	}
	return err
}

func (b *btree) Delete(key uint64) error {
	if b.RootID == 0 {
		return nil
	}
	root := b.Cache.Get(b.RootID)
	deleted, err := root.Delete(b.Cache, key)
	if deleted {
		b.Size--
	}
	b.RootID = root.GetNewRootID()
	if b.RootID != 0 {
		b.Cache.Get(b.RootID).SetParentID(0)
	}
	return err
}

func (b *btree) Print() {
	if b.RootID == 0 {
		fmt.Println("Empty tree")
		return
	}
	b.Cache.Get(b.RootID).Print(b.Cache, 0)
}


func (b *btree) GetRange(start, end uint64) []uint64 {
	if b.RootID == 0 {
		return nil
	}
	res := make([]uint64, 0)
	res = b.Cache.Get(b.RootID).GetRange(b.Cache, start, end, res)
	return res
}




