package node


type NodeCache interface {
	Get(id uint64) BTreeNode
	Register(n BTreeNode) error
}

type NodeCacheImpl struct {
	Nodes map[uint64]BTreeNode
	size uint64
}

func NewNodeCacheImpl() *NodeCacheImpl {
	return &NodeCacheImpl{
		Nodes: make(map[uint64]BTreeNode),
		size: 0,
	}
}

func (c *NodeCacheImpl) Get(id uint64) BTreeNode {
	if n, ok := c.Nodes[id]; ok {
		return n
	}

	// TODO: read from disk

	// TODO: deserialize
	return nil
}

func (c *NodeCacheImpl) Register(n BTreeNode) error {
	c.size++
	c.Nodes[c.size] = n
	n.SetID(c.size)
	return nil
}



