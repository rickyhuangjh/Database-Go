package serialization


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

const pageSize = 4096


const parentIDSize = 64
const prevIDSize = 64
const nextIDSize = 64
const childIDSize = 64

const numKeySize = 32
const numChildrenSize = 32
const numValSize = 32

const keySize = 64
const valSize = 64

const internalHeaderSize = parentIDSize + numKeySize + numChildrenSize

const leafHeaderSize = parentIDSize + prevIDSize + nextIDSize + numKeySize + numValSize




