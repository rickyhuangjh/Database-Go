package node

import (
	"btree/utils"
	"fmt"
	"strings"
)


func NewLeafNode() *LeafNode {
	return &LeafNode{
		ID: 0,
		Keys:     make([]uint64, 0, utils.OptimalLeafOrder + 1),
		Vals:     make([]uint64, 0, utils.OptimalLeafOrder + 1),
		ParentID:   nil,
		NextID:     nil,
		PrevID:     nil,
	}
}


// returns the index of where the KVP would go
func (n *LeafNode) find(key uint64) (int, bool) {
	for i, curKey := range n.Keys {
		if key == curKey {
			return i, true
		} else if key < curKey {
			return i, false
		}
	}
	return len(n.Keys), false
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

func (n *LeafNode) Traverse(cache NodeCache, res []uint64) []uint64 {
	res = append(res, n.Vals...)
	if n.NextID == 0 {
		return res
	}
	return cache.Get(n.NextID).Traverse(res)
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
	idx, wasFound := n.find(key)
	if !wasFound {
		return false, nil
	}
	old := n.Keys[0]
	n.Keys = utils.Delete(n.Keys, idx)
	n.Vals = utils.Delete(n.Vals, idx)
	
	err := n.merge()
	parentNode := c.Get(n.ParentID).(*InternalNode)
	if parentNode != nil && len(n.Keys) > 0 {
		parentNode.replaceKey(old, n.Keys[0])
	}
	return true, err
}


func (n *LeafNode) GetNewRootID() uint64 {
	if len(n.Keys) == 0 {
		return 0
	}
	return n.ID
}

func (n *LeafNode) Print(level int) {
    indent := strings.Repeat("    ", level)
    fmt.Printf("%sLeaf Node: keys=%v", indent, n.Keys)
}

func (n *LeafNode) Verify() (uint64, uint64) {
	return n.Keys[0], n.Keys[len(n.Keys)-1]
}

func (n *LeafNode) split(c NodeCache) error {
	if len(n.Vals) <= utils.LeafOrder {
		return nil
	}

	siblingNode := NewLeafNode()
	siblingID, err := c.Register(siblingNode)
	if err != nil {
		return err
	}

	if n.ParentID == 0 {
		parentID, err := c.Register(NewInternalNode())
		if err != nil {
			return err
		}
		n.ParentID = parentID
		c.Get(n.ParentID).insertChild(0, n)
	}
	parentNode := c.Get(n.ParentID).(*InternalNode)
	siblingNode.ParentID = n.ParentID

	mid := len(n.Keys) / 2

	idx, _ := parentNode.find(n.Keys[mid])
	parentNode.Keys = utils.Insert(parentNode.Keys, idx, n.Keys[mid])
	parentNode.ChildIDs = utils.Insert(parentNode.ChildIDs, idx+1, siblingID)

	siblingNode.Keys = make([]uint64, len(n.Keys[mid:]), utils.LeafOrder)
	copy(siblingNode.Keys, n.Keys[mid:])
	n.Keys = n.Keys[:mid]

	siblingNode.Vals = make([]uint64, len(n.Vals[mid:]), utils.LeafOrder)
	copy(siblingNode.Vals, n.Vals[mid:])
	n.Vals = n.Vals[:mid]

	siblingNode.NextID = n.NextID

	if n.NextID != 0 {
		nextNode := c.Get(n.NextID).(*LeafNode)
		c.Get(nextNode.ID).(*LeafNode).PrevID = siblingNode.ID
	}
	n.NextID = siblingNode.ID
	siblingNode.PrevID = n.ID
	return parentNode.split()
}

func (n *LeafNode) merge(c NodeCache) error {
	if n.ParentID == 0 || len(n.Vals) >= int(utils.LeafOrder)/2 {
		return nil
	}
	parentNode := c.Get(n.ParentID).(*InternalNode)
	nextNode := c.Get(n.NextID).(*LeafNode)
	prevNode := c.Get(n.PrevID).(*LeafNode)
	if nextNode != nil && (c.Get(nextNode.ParentID) == parentNode) &&
		len(nextNode.Keys) > int(utils.LeafOrder+1)/2 {
		// steal from next
		parentNode.replaceKey(nextNode.Keys[0], nextNode.Keys[1])
		n.Keys = append(n.Keys, nextNode.Keys[0])
		n.Vals = append(n.Vals, nextNode.Vals[0])
		nextNode.Keys = nextNode.Keys[1:]
		nextNode.Vals = nextNode.Vals[1:]
	} else if prevNode != nil && prevNode.Parent == n.Parent &&
		len(prevNode.Keys) > int(utils.LeafOrder+1)/2 {
		// steal from prev
		n.Keys = utils.Insert(n.Keys, 0, prevNode.Keys[len(prevNode.Keys)-1])
		n.Vals = utils.Insert(n.Vals, 0, prevNode.Vals[len(prevNode.Vals)-1])
		prevNode.Keys = prevNode.Keys[:len(prevNode.Keys)-1]
		prevNode.Vals = prevNode.Vals[:len(prevNode.Vals)-1]
	} else if prevNode != nil && prevNode.Parent == parentNode {
		// merge with prev
		prevNode.Keys = append(prevNode.Keys, n.Keys...)
		prevNode.Vals = append(prevNode.Vals, n.Vals...)
		ourIdx, _ := parentNode.findChildIdx(n.ID)
		parentNode.Keys = utils.Delete(parentNode.Keys, ourIdx-1)
		parentNode.ChildIDs = utils.Delete(parentNode.ChildIDs, ourIdx)
		prevNode.NextID = n.NextID
		if nextNode != nil {
			nextNode.PrevID = n.PrevID
		}
	} else {
		// merge with next
		n.Keys = append(n.Keys, nextNode.Keys...)
		n.Vals = append(n.Vals, nextNode.Vals...)
		ourIdx, _ := parentNode.findChildIdx(n.ID)
		parentNode.Keys = utils.Delete(parentNode.Keys, ourIdx)
		parentNode.ChildIDs = utils.Delete(parentNode.ChildIDs, ourIdx+1)
		
		nextNextNode := c.Get(nextNode.NextID).(*LeafNode)
		if nextNextNode != nil {
			nextNextNode.PrevID = n.ID
		}
		n.NextID = nextNextNode.ID
	}
	return parentNode.merge()
}
