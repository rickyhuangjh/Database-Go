package node

import (
	"btree/io"
)

type NodeCache interface {
	Get(id uint64) BTreeNode
	Register(n BTreeNode) error
}

type nodeCache struct {
	manager io.StorageManager
	nodes map[uint64]BTreeNode
	size uint64
}

func NewNodeCache(manager io.StorageManager) NodeCache {
	return &nodeCache{
		manager: manager,
		nodes: make(map[uint64]BTreeNode),
		size: 0,
	}
}



func (c *nodeCache) Get(id uint64) BTreeNode {
	if id == 0 {
		return nil
	} else if n, ok := c.nodes[id]; ok {
		return n
	}

	// TODO: read from disk

	// TODO: deserialize
	return nil
}

func (c *nodeCache) Register(n BTreeNode) error {
	c.size++
	c.nodes[c.size] = n
	n.SetID(c.size)
	return nil
}



