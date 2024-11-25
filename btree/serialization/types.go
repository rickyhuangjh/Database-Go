package serialization

import "btree/node"

type BTreeNodeSerialize interface {
	Deserialize() node.BTreeNode
}

type InternalNodeSerialized struct {
	ParentID    uint64 // MSB is 1
	NumKeys     uint32
	NumChildren uint32
	Keys        []uint64
	ChildrenIDs    []uint64
}

type LeafNodeSerialized struct {
	ParentID uint64 // MSB is 0
	PrevID   uint64
	NextID   uint64
	NumKeys  uint32
	NumVals  uint32
	Keys     []uint64
	Vals     []uint64 // max size = 4096 - header size
}