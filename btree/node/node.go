package node


type BTreeNode interface {
	GetID() uint64
	SetID(id uint64)
	GetParentID() uint64
	SetParentID(id uint64)
	Get(c NodeCache, key uint64) (uint64, bool)
	GetRange(c NodeCache, start, end uint64, res []uint64) []uint64
	Traverse(c NodeCache, res []uint64) []uint64
	Set(c NodeCache, key uint64, val uint64) (bool, error)
	Delete(c NodeCache, key uint64) (bool, error)
	Print(c NodeCache, level int)
	GetNewRootID() uint64
	split(c NodeCache) error
	merge(c NodeCache) error
}

type InternalNode struct {
	ID uint64
	ParentID uint64
	Keys     []uint64
	ChildIDs []uint64
}

type LeafNode struct {
	ID uint64 // Left-most bit is 1
	ParentID uint64
	NextID uint64
	PrevID   uint64
	Keys   []uint64
	Vals   []uint64
}

const internalHeaderSize = 16
const leafHeaderSize = 32

const keySize = 8
const childIDSize = 8
const valSize = 8


const internalOrder = (4096 - internalHeaderSize) / (keySize + childIDSize)
const leafOrder = (4096 - leafHeaderSize) / (keySize + valSize)


