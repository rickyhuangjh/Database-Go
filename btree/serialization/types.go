package serialization

type InternalNodeSerialized struct {
	Order       int32 // negative
	ParentID    uint32
	NumKeys     uint32
	NumChildren uint32
	Keys        []int32
	Children    []uint32
}

type LeafNodeSerialized struct {
	Order    int32 // positive
	ParentID uint32
	PrevID   uint32
	NextID   uint32
	NumKeys  uint32
	NumVals  uint32
	Keys     []int32
	Vals     [][]byte // max size = 4096 - header size
}

const PageSize = 4096

const OrderSize = 32
const ParentIDSize = 32

const PrevIDSize = 32
const NextIDSize = 32

const NumKeySize = 32
const NumChildrenSize = 32
const NumValSize = 32

const KeySize = 32

const InternalHeaderSize = OrderSize + ParentIDSize + NumKeySize + NumChildrenSize
const LeafHeaderSize = OrderSize + ParentIDSize + PrevIDSize + NextIDSize + NumKeySize + NumValSize
