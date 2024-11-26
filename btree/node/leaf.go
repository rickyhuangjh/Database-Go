package node

import (
	"btree/utils"
	"fmt"
	"strings"
)


func NewLeafNode() *LeafNode {
	return &LeafNode{
		ID: 0,
		Keys:     make([]uint64, 0, leafOrder + 1),
		Vals:     make([]uint64, 0, leafOrder + 1),
		ParentID:   0,
		NextID:     0,
		PrevID:     0,
	}
}

func (n *LeafNode) GetID() uint64 {
	return n.ID
}

func (n *LeafNode) SetID(id uint64) {
	n.ID = id
}

func (n *LeafNode) GetParentID() uint64 {
	return n.ParentID
}

func (n *LeafNode) SetParentID(id uint64) {
	n.ParentID = id
}

// returns the index of where the KVP would go
func (n *LeafNode) find(key uint64) (int, bool) {
	res := len(n.Keys)
	for left, right := 0, len(n.Keys) - 1; left <= right; {
		mid := (left + right) / 2
		if (n.Keys[mid] == key) {
			return mid, true
		} else if (n.Keys[mid] < key) {
			left = mid + 1
		} else {
			res = mid
			right = mid - 1
		}
	}
	return res, false
}

func (n *LeafNode) Get(cache NodeCache, key uint64) (uint64, bool) {
	valIdx, found := n.find(key)
	if found {
		return n.Vals[valIdx], true
	}
	return 0, false
}

func (n *LeafNode) GetRange(cache NodeCache, start, end uint64, res []uint64) []uint64 {
	startIdx, _ := n.find(start)
	endIdx, _ := n.find(end)
	res = append(res, n.Vals[startIdx:endIdx]...)
	if n.NextID > 0 && endIdx == len(n.Keys) {
		return cache.Get(n.NextID).GetRange(cache, start, end, res)
	}
	return res
}

func (n *LeafNode) Traverse(c NodeCache, res []uint64) []uint64 {
	res = append(res, n.Vals...)
	if n.NextID == 0 {
		return res
	}
	return c.Get(n.NextID).Traverse(c, res)
}

func (n *LeafNode) Set(c NodeCache, key uint64, val uint64) (bool, error) {
	idx, found := n.find(key)
	if found {
		n.Vals[idx] = val
	} else {
		n.Keys = utils.Insert(n.Keys, idx, key)
		n.Vals = utils.Insert(n.Vals, idx, val)
	}
	return !found, n.split(c)
}

func (n *LeafNode) Delete(c NodeCache, key uint64) (bool, error) {
	idx, found := n.find(key)
	if !found {
		return false, nil
	}
	old := n.Keys[0]
	n.Keys = utils.Delete(n.Keys, idx)
	n.Vals = utils.Delete(n.Vals, idx)
	
	err := n.merge(c)
	parent := c.Get(n.ParentID).(*InternalNode)
	if parent != nil && len(n.Keys) > 0 {
		parent.replaceKey(c, old, n.Keys[0])
	}
	return true, err
}


func (n *LeafNode) split(c NodeCache) error {
	if len(n.Vals) <= leafOrder {
		return nil
	}

	sibling := NewLeafNode()
	err := c.Register(sibling)
	if err != nil {
		return err
	}

	var parent *InternalNode
	if n.ParentID == 0 {
		parent = NewInternalNode()
		err := c.Register(parent)
		if err != nil {
			return err
		}
		n.ParentID = parent.ID
		parent.ChildIDs = utils.Insert(parent.ChildIDs, 0, n.ID)
	} else {
		parent = c.Get(n.ParentID).(*InternalNode)
	}
	sibling.ParentID = n.ParentID

	mid := len(n.Keys) / 2

	ourIdx, _ := parent.findChildIdx(n.ID)
	// fmt.Println(parent.ChildIDs)
	// fmt.Println(n.ID)
	parent.Keys = utils.Insert(parent.Keys, ourIdx, n.Keys[mid])
	parent.ChildIDs = utils.Insert(parent.ChildIDs, ourIdx+1, sibling.ID)

	sibling.Keys = make([]uint64, len(n.Keys[mid:]), leafOrder+1)
	copy(sibling.Keys, n.Keys[mid:])
	n.Keys = n.Keys[:mid]

	sibling.Vals = make([]uint64, len(n.Vals[mid:]), leafOrder+1)
	copy(sibling.Vals, n.Vals[mid:])
	n.Vals = n.Vals[:mid]

	sibling.NextID = n.NextID

	if n.NextID != 0 {
		next := c.Get(n.NextID).(*LeafNode)
		c.Get(next.ID).(*LeafNode).PrevID = sibling.ID
	}
	n.NextID = sibling.ID
	sibling.PrevID = n.ID
	return parent.split(c)
}

func (n *LeafNode) merge(c NodeCache) error {
	if n.ParentID == 0 || len(n.Vals) >= int(leafOrder)/2 {
		return nil
	}
	parent := c.Get(n.ParentID).(*InternalNode)
	next := c.Get(n.NextID).(*LeafNode)
	prev := c.Get(n.PrevID).(*LeafNode)
	if next != nil && (c.Get(next.ParentID) == parent) &&
		len(next.Keys) > int(leafOrder+1)/2 {
		// steal from next
		parent.replaceKey(c, next.Keys[0], next.Keys[1])
		n.Keys = append(n.Keys, next.Keys[0])
		n.Vals = append(n.Vals, next.Vals[0])
		next.Keys = next.Keys[1:]
		next.Vals = next.Vals[1:]
	} else if prev != nil && prev.ParentID == n.ParentID &&
		len(prev.Keys) > (leafOrder+1)/2 {
		// steal from prev
		n.Keys = utils.Insert(n.Keys, 0, prev.Keys[len(prev.Keys)-1])
		n.Vals = utils.Insert(n.Vals, 0, prev.Vals[len(prev.Vals)-1])
		prev.Keys = prev.Keys[:len(prev.Keys)-1]
		prev.Vals = prev.Vals[:len(prev.Vals)-1]
	} else if prev != nil && prev.ParentID == n.ParentID {
		// merge with prev
		prev.Keys = append(prev.Keys, n.Keys...)
		prev.Vals = append(prev.Vals, n.Vals...)
		ourIdx, _ := parent.findChildIdx(n.ID)
		parent.Keys = utils.Delete(parent.Keys, ourIdx-1)
		parent.ChildIDs = utils.Delete(parent.ChildIDs, ourIdx)
		prev.NextID = n.NextID
		if next != nil {
			next.PrevID = n.PrevID
		}
	} else {
		// merge with next
		n.Keys = append(n.Keys, next.Keys...)
		n.Vals = append(n.Vals, next.Vals...)
		ourIdx, _ := parent.findChildIdx(n.ID)
		parent.Keys = utils.Delete(parent.Keys, ourIdx)
		parent.ChildIDs = utils.Delete(parent.ChildIDs, ourIdx+1)
		
		nextNext := c.Get(next.NextID).(*LeafNode)
		if nextNext != nil {
			nextNext.PrevID = n.ID
		}
		n.NextID = nextNext.ID
	}
	return parent.merge(c)
}

func (n *LeafNode) GetNewRootID() uint64 {
	if len(n.Vals) == 0 {
		return 0
	}
	return n.ID
}

func (n *LeafNode) Print(c NodeCache, level int) {
    indent := strings.Repeat("    ", level)
    fmt.Printf("%sLeaf Node: keys=%v\n", indent, n.Keys)
}

