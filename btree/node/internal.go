package node

import (
	"btree/utils"
	"fmt"
	"strings"
)

func NewInternalNode() *InternalNode {
	return &InternalNode{
		ParentID:   0,
		Keys:     make([]uint64, 0, internalOrder),
		ChildIDs: make([]uint64, 0, internalOrder + 1),
	}
}


func (n *InternalNode) GetID() uint64 {
	return n.ID
}

func (n *InternalNode) SetID(id uint64) {
	n.ID = id
}

func (n *InternalNode) GetParentID() uint64 {
	return n.ParentID
}

func (n *InternalNode) SetParentID(id uint64) {
	n.ParentID = id
}

func (n *InternalNode) find(key uint64) (int, bool) {
	res := len(n.Keys)
	for left, right := 0, len(n.Keys) - 1; left <= right; {
		mid := (left + right) / 2
		if n.Keys[mid] <= key {
			res = mid
			left = mid + 1
		} else {
			right = mid - 1;
		}
	}
	return res, false
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
	if len(n.ChildIDs) <= internalOrder {
		return nil
	}

	sibling := NewInternalNode()
	c.Register(sibling)

	var parent *InternalNode
	if c.IsValid(n.ParentID) {
		parent = c.Get(n.ParentID).(*InternalNode)
	} else {
		parent = NewInternalNode()
		err := c.Register(parent)
		if err != nil {
			return err
		}
		n.ParentID = parent.ID
		parent.ChildIDs = utils.Insert(parent.ChildIDs, 0, n.ID)
	}

	sibling.ParentID= n.ParentID

	mid := len(n.Keys) / 2

	idx, _ := parent.find(n.Keys[mid])
	parent.Keys = utils.Insert(parent.Keys, idx, n.Keys[mid])
	parent.ChildIDs = utils.Insert(parent.ChildIDs, idx+1, sibling.ID)

	sibling.Keys = make([]uint64, len(n.Keys[mid+1:]), internalOrder)
	copy(sibling.Keys, n.Keys[mid+1:])
	n.Keys = n.Keys[:mid]

	sibling.ChildIDs = make([]uint64, len(n.ChildIDs[mid+1:]), internalOrder+1)
	copy(sibling.ChildIDs, n.ChildIDs[mid+1:])
	n.ChildIDs = n.ChildIDs[:mid+1]

	for _, childID := range sibling.ChildIDs {
		c.Get(childID).SetParentID(sibling.ID)
	}

	return parent.split(c)
}

func (n *InternalNode) merge(c NodeCache) error {

	if c.IsValid(n.ParentID) || len(n.ChildIDs) >= (internalOrder + 1)/2 {
		return nil
	}

	parent := c.Get(n.ParentID).(*InternalNode)
	ourIdx, _ := parent.findChildIdx(n.ID)

	// Try to steal from siblings first
	if ourIdx > 0 &&
		len(c.Get(parent.ChildIDs[ourIdx-1]).(*InternalNode).ChildIDs) >
		(internalOrder + 1)/2 {
		// Steal from left sibling
		leftSibling := c.Get(parent.ChildIDs[ourIdx-1]).(*InternalNode)
		n.Keys = utils.Insert(n.Keys, 0, parent.Keys[ourIdx-1])
		parent.Keys[ourIdx-1] = leftSibling.Keys[len(leftSibling.Keys)-1]
		leftSibling.Keys = leftSibling.Keys[:len(leftSibling.Keys)-1]
		n.ChildIDs = utils.Insert(n.ChildIDs, 0,
			leftSibling.ChildIDs[len(leftSibling.ChildIDs)-1])
		leftSibling.ChildIDs = leftSibling.ChildIDs[:len(leftSibling.ChildIDs)-1]
		c.Get(n.ChildIDs[0]).SetParentID(n.ID)

		return nil
	} else if ourIdx < len(parent.ChildIDs)-1 &&
		len(c.Get(parent.ChildIDs[ourIdx+1]).(*InternalNode).ChildIDs) >
		int(internalOrder + 1)/2 {
		// Steal from right sibling
		rightSibling := c.Get(parent.ChildIDs[ourIdx+1]).(*InternalNode)
		n.Keys = append(n.Keys, parent.Keys[ourIdx])
		parent.Keys[ourIdx] = rightSibling.Keys[0]
		rightSibling.Keys = rightSibling.Keys[1:]
		rightSibling.ChildIDs = rightSibling.ChildIDs[1:]
		c.Get(n.ChildIDs[len(n.ChildIDs)-1]).SetParentID(n.ID)

		return nil
	}

	if ourIdx > 0 {
		// Merge with left sibling
		leftSibling := c.Get(parent.ChildIDs[ourIdx-1]).(*InternalNode)
		leftSibling.Keys = append(leftSibling.Keys, parent.Keys[ourIdx-1])
		leftSibling.Keys = append(leftSibling.Keys, n.Keys...)
		leftSibling.ChildIDs = append(leftSibling.ChildIDs, n.ChildIDs...)
		for _, childID := range n.ChildIDs {
			c.Get(childID).SetParentID(leftSibling.ID)
		}
		parent.Keys = utils.Delete(parent.Keys, ourIdx-1)
		parent.ChildIDs = utils.Delete(parent.ChildIDs, ourIdx)

		return parent.merge(c)
	} else {
		// Merge with right sibling
		rightSibling := c.Get(parent.ChildIDs[ourIdx+1]).(*InternalNode)
		n.Keys = append(n.Keys, parent.Keys[ourIdx])
		n.Keys = append(n.Keys, rightSibling.Keys...)
		n.ChildIDs = append(n.ChildIDs, rightSibling.ChildIDs...)
		for _, childID := range rightSibling.ChildIDs {
			c.Get(childID).SetParentID(n.ID)
		}
		parent.Keys = utils.Delete(parent.Keys, ourIdx)
		parent.ChildIDs = utils.Delete(parent.ChildIDs, ourIdx+1)

		return parent.merge(c)
	}
}

func (n *InternalNode) GetNewRootID() uint64 {
	if len(n.Keys) == 0 {
		return n.ChildIDs[0]
	}
	return n.ID
}

func (n *InternalNode) Print(c NodeCache, level int) {
	indent := strings.Repeat("    ", level)
	fmt.Printf("%sInternal Node: keys=%v\n", indent, n.Keys)
	for _, childID := range n.ChildIDs {
		c.Get(childID).Print(c, level + 1)
	}
}

