package node

import (
	"btree/utils"
	"fmt"
	"strings"
)

func NewInternalNode() *InternalNode {
	return &InternalNode{
		ParentID:   0,
		Keys:     make([]uint64, 0, utils.InternalOrder),
		ChildIDs: make([]uint64, 0, utils.InternalOrder + 1),
	}
}



func (n *InternalNode) find(key uint64) (int, bool) {
	for i, curKey := range n.Keys {
		if key < curKey {
			return i, true
		}
	}
	return len(n.Keys), false
}

func (n *InternalNode) findChildIdx(childID uint64) (int, bool) {
	for i, curChildID := range n.ChildIDs {
		if curChildID == childID {
			return i, true
		}
	}
	return -1, false
}

func (n *InternalNode) Get(c NodeCache, key uint64) (uint64, bool) {
	idx, _ := n.find(key)
	return c.Get(n.ChildIDs[idx]).Get(c, key)
}

func (n *InternalNode) GetRange(c NodeCache,
	start, end uint64, res []uint64) []uint64 {
	idx, _ := n.find(start)
	return c.Get(n.ChildIDs[idx]).GetRange(c, start, end, res)
}

func (n *InternalNode) Traverse(c NodeCache, res []uint64) []uint64 {
	return c.Get(n.ChildIDs[0]).Traverse(c, res)
}

func (n *InternalNode) Set(c NodeCache, key uint64, val uint64) (bool, error) {
	idx, _ := n.find(key)
	return c.Get(n.ChildIDs[idx]).Set(c, key, val)
}

func (n *InternalNode) Delete(c NodeCache, key uint64) (bool, error) {
	idx, _ := n.find(key)
	return c.Get(n.ChildIDs[idx]).Delete(c, key)
}

func (n *InternalNode) replaceKey(c NodeCache, oldKey uint64, newKey uint64) {
	if oldKey == newKey {
		return
	}


	if len(n.Keys) == 0 || oldKey < n.Keys[0] {
		parentNode := c.Get(n.ParentID)
		if parentNode != nil {
			parentNode.(*InternalNode).replaceKey(c, oldKey, newKey)
		}
		return
	}
	for i, curKey := range n.Keys {
		if oldKey == curKey {
			n.Keys[i] = newKey
			return
		}
	}
}

func (n *InternalNode) split(c NodeCache) error {
	if len(n.ChildIDs) <= utils.InternalOrder {
		return nil
	}


	parentNode := c.Get(n.ParentID)
	

	siblingNode := NewInternalNode()
	c.Register(siblingNode)
	if n.ParentID == 0 {
		parentNode := newInternalNode()
		c.Register(parentNode)
		parentNode.ChildIDs = insert(parentNode.ChildIDs, 0, n.ID)
	}
	siblingNode.Parent = n.Parent

	mid := len(n.Keys) / 2

	idx, _ := n.Parent.find(n.Keys[mid])
	n.Parent.Keys = utils.Insert(n.Parent.Keys, idx, n.Keys[mid])
	n.Parent.insertChild(idx+1, siblingNode)

	siblingNode.Keys = make([]uint64, len(n.Keys[mid+1:]), n.Order+1)
	copy(siblingNode.Keys, n.Keys[mid+1:])
	n.Keys = n.Keys[:mid]

	siblingNode.Children = make([]BTreeNode, len(n.Children[mid+1:]), n.Order+1)
	copy(siblingNode.Children, n.Children[mid+1:])
	n.Children = n.Children[:mid+1]

	for _, child := range siblingNode.Children {
		child.SetParent(siblingNode)
	}

	return n.Parent.split()
}

func (n *InternalNode) merge(c nodeCache) error {
	if n.Parent == nil || len(n.Children) >= int(n.Order + 1)/2 {
		return nil
	}

	ourIdx := 0
	for i, child := range n.Parent.Children {
		if child == n {
			ourIdx = i
			break
		}
	}

	// Try to steal from siblings first
	if ourIdx > 0 &&
		len(n.Parent.Children[ourIdx-1].(*InternalNode).Children) > int(n.Order + 1)/2 {
		// Steal from left sibling
		leftSibling := n.Parent.Children[ourIdx-1].(*InternalNode)
		n.Keys = utils.Insert(n.Keys, 0, n.Parent.Keys[ourIdx-1])
		n.Parent.Keys[ourIdx-1] = leftSibling.Keys[len(leftSibling.Keys)-1]
		leftSibling.Keys = leftSibling.Keys[:len(leftSibling.Keys)-1]
		n.Children = utils.Insert(n.Children, 0, leftSibling.Children[len(leftSibling.Children)-1])
		leftSibling.Children = leftSibling.Children[:len(leftSibling.Children)-1]
		n.Children[0].SetParent(n)

		return nil
	} else if ourIdx < len(n.Parent.Children)-1 &&
		len(n.Parent.Children[ourIdx+1].(*InternalNode).Children) > int(n.Order + 1)/2 {
		// Steal from right sibling
		rightSibling := n.Parent.Children[ourIdx+1].(*InternalNode)
		n.Keys = append(n.Keys, n.Parent.Keys[ourIdx])
		n.Parent.Keys[ourIdx] = rightSibling.Keys[0]
		rightSibling.Keys = rightSibling.Keys[1:]
		rightSibling.Children = rightSibling.Children[1:]
		n.Children[len(n.Children)-1].SetParent(n)

		return nil
	}

	if ourIdx > 0 {
		// Merge with left sibling
		leftSibling := n.Parent.Children[ourIdx-1].(*InternalNode)
		leftSibling.Keys = append(leftSibling.Keys, n.Parent.Keys[ourIdx-1])
		leftSibling.Keys = append(leftSibling.Keys, n.Keys...)
		leftSibling.Children = append(leftSibling.Children, n.Children...)
		for _, child := range n.Children {
			child.SetParent(leftSibling)
		}
		n.Parent.Keys = utils.Delete(n.Parent.Keys, ourIdx-1)
		n.Parent.Children = utils.Delete(n.Parent.Children, ourIdx)

		return n.Parent.merge()
	} else {
		// Merge with right sibling
		rightSibling := n.Parent.Children[ourIdx+1].(*InternalNode)
		n.Keys = append(n.Keys, n.Parent.Keys[ourIdx])
		n.Keys = append(n.Keys, rightSibling.Keys...)
		n.Children = append(n.Children, rightSibling.Children...)
		for _, child := range rightSibling.Children {
			child.SetParent(n)
		}
		n.Parent.Keys = utils.Delete(n.Parent.Keys, ourIdx)
		n.Parent.Children = utils.Delete(n.Parent.Children, ourIdx+1)

		return n.Parent.merge()
	}
}

func (n *InternalNode) GetNewRoot() BTreeNode {
	if len(n.Keys) == 0 {
		return n.Children[0]
	}
	return n
}

func (n *InternalNode) Print(level int) {
	indent := strings.Repeat("    ", level)
	fmt.Printf("%sInternal Node: keys=%v\n", indent, n.Keys)
	for _, child := range n.Children {
		child.Print(level + 1)
	}
}

func (n *InternalNode) Verify() (uint64, uint64) {
	for i := range n.Keys {
		_, wasFound := n.Children[i+1].Get(n.Keys[i])
		if !wasFound {
			panic(fmt.Sprintf("key %v not found in node %v\n", n.Keys[i], n))
		}
	}
	a, _ := n.Children[0].Verify()
	_, b := n.Children[len(n.Children)-1].Verify()
	for i := 0; i < len(n.Keys); i++ {
		x, _ := n.Children[i].Verify()
		if x >= n.Keys[i] {
			panic("uh oh\n")
		}
		y, _ := n.Children[i+1].Verify()
		if y != n.Keys[i] {
			panic("uh oh\n")
		}
	}
	for _, child := range n.Children {
		child.Verify()
	}
	return a, b
}
