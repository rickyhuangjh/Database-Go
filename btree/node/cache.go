package node

import (
	// "btree/io"
	"errors"
)

const maxType = (1 << 63) - 1

type NodeCache interface {
	IsValid(id uint64) bool
	Get(id uint64) BTreeNode
	Register(n BTreeNode) error
}

type nodeCache struct {
	// manager io.StorageManager
	nodes map[uint64]BTreeNode
	size uint64
	internalNodes uint64
	leafNodes uint64
}

func NewNodeCache() NodeCache {
	return &nodeCache{
		// manager: manager,
		nodes: make(map[uint64]BTreeNode),
		size: 0,
	}
}


func (c *nodeCache) IsValid(id uint64) bool {
	return id != 0 && id != (1 << 63)
}

func (c *nodeCache) Get(id uint64) BTreeNode {
	if id == 0 {
		return nil
	} else if n, ok := c.nodes[id]; ok {
		Deserialize(Serialize(n))
		return n
	}

	// TODO: read from disk

	// TODO: deserialize
	return nil
}

func (c *nodeCache) Register(n BTreeNode) error {
	var id uint64
	switch n.(type) {
	case *InternalNode:
		if c.internalNodes == maxType {
			return errors.New("Too many InternalNodes")
		}
		c.internalNodes++
		id = c.internalNodes
	default:
		if c.leafNodes == maxType {
			return errors.New("Too many LeafNodes")
		}
		c.leafNodes++
		id = c.leafNodes | (1 << 63)
	}
	c.size++
	c.nodes[id] = n
	n.SetID(id)
	return nil
}


