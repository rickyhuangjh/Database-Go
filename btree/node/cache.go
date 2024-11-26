package node

type NodeCache interface {
	Get(id uint64) BTreeNode
	Register(BTreeNode) error
	Update(uint64, BTreeNode) error
}

type NodeCacheImpl struct {
	nodes map[uint64]BTreeNode
}

func (c *NodeCacheImpl) Get(id uint64) (BTreeNode, bool) {
	if n, ok := c.nodes[id]; ok {
		return n, true
	}

	// TODO: read from disk

	// TODO: deserialize
	return nil, false
}



