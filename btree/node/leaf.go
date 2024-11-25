package node

import (
	"btree/utils"
	"fmt"
	"strings"
)


func NewLeafNode(order uint64) *LeafNode {
	return &LeafNode{
		Order:    order,
		Keys:     make([]uint64, 0, order),
		Vals:     make([]V, 0, order),
		Parent:   nil,
		Next:     nil,
		Prev:     nil,
	}
}

func (n *LeafNode) GetParent() *InternalNode {
	return n.Parent
}

func (n *LeafNode) SetParent(parent *InternalNode) error {
	n.Parent = parent
	return nil
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

func (n *LeafNode) Get(key uint64) (V, bool) {
	idx, wasFound := n.find(key)
	if wasFound {
		return n.Vals[idx], true
	}
	var zero V
	return zero, false
}

func (n *LeafNode) GetRange(start, end uint64, res []V) []V {
	startIdx, _ := n.find(start)
	endIdx, _ := n.find(end)
	res = append(res, n.Vals[startIdx:endIdx]...)
	if n.Next != nil && endIdx == len(n.Keys) {
		return n.Next.GetRange(start, end, res)
	}
	return res
}

func (n *LeafNode) Traverse(res []V) []V {
	res = append(res, n.Vals...)
	if n.Next == nil {
		return res
	}
	return n.Next.Traverse(res)
}

func (n *LeafNode) Set(key uint64, val V) (bool, error) {
	idx, wasFound := n.find(key)
	if wasFound {
		n.Vals[idx] = val
	} else {
		n.Keys = utils.Insert(n.Keys, idx, key)
		n.Vals = utils.Insert(n.Vals, idx, val)
	}
	return !wasFound, n.split()
}

func (n *LeafNode) Delete(key uint64) (bool, error) {
	idx, wasFound := n.find(key)
	if !wasFound {
		return false, nil
	}
	old := n.Keys[0]
	n.Keys = utils.Delete(n.Keys, idx)
	n.Vals = utils.Delete(n.Vals, idx)
	
	err := n.merge()
	if n.Parent != nil && len(n.Keys) > 0 {
		n.Parent.replaceKey(old, n.Keys[0])
	}
	return true, err
}


func (n *LeafNode) GetNewRoot() BTreeNode {
	if len(n.Keys) == 0 {
		return nil
	}
	return n
}

func (n *LeafNode) Print(level int) {
    indent := strings.Repeat("    ", level)
	next := n.Next
	prev := n.Prev
	nextKeys := []uint64{}
	if next != nil {
		nextKeys = next.Keys
	}
	prevKeys := []uint64{}
	if prev != nil {
		prevKeys = prev.Keys
	}
    fmt.Printf("%sLeaf Node: keys=%v, next=%v, prev=%v\n", indent, n.Keys, nextKeys, prevKeys)
}

func (n *LeafNode) Verify() (uint64, uint64) {
	return n.Keys[0], n.Keys[len(n.Keys)-1]
}

func (n *LeafNode) split() error {
	if len(n.Keys) < int(n.Order) {
		return nil
	}

	siblingNode := NewLeafNode(n.Order)
	if n.Parent == nil {
		n.Parent = newInternalNode(n.Order)
		n.Parent.insertChild(0, n)
	}
	siblingNode.Parent = n.Parent

	mid := len(n.Keys) / 2

	idx, _ := n.Parent.find(n.Keys[mid])
	n.Parent.Keys = utils.Insert(n.Parent.Keys, idx, n.Keys[mid])
	n.Parent.insertChild(idx+1, siblingNode)

	siblingNode.Keys = make([]uint64, len(n.Keys[mid:]), n.Order)
	copy(siblingNode.Keys, n.Keys[mid:])
	n.Keys = n.Keys[:mid]

	siblingNode.Vals = make([]V, len(n.Vals[mid:]), n.Order)
	copy(siblingNode.Vals, n.Vals[mid:])
	n.Vals = n.Vals[:mid]

	siblingNode.Next = n.Next
	if n.Next != nil {
		n.Next.Prev = siblingNode
	}
	n.Next = siblingNode
	siblingNode.Prev = n
	return n.Parent.split()
}

func (n *LeafNode) merge() error {
	if n.Parent == nil || len(n.Keys) >= int(n.Order-1)/2 {
		return nil
	}
	if n.Next != nil && n.Next.Parent == n.Parent &&
		len(n.Next.Keys) > int(n.Order-1)/2 {
		// steal from next
		n.Parent.replaceKey(n.Next.Keys[0], n.Next.Keys[1])
		n.Keys = append(n.Keys, n.Next.Keys[0])
		n.Vals = append(n.Vals, n.Next.Vals[0])
		n.Next.Keys = n.Next.Keys[1:]
		n.Next.Vals = n.Next.Vals[1:]
	} else if n.Prev != nil && n.Prev.Parent == n.Parent &&
		len(n.Prev.Keys) > int(n.Order-1)/2 {
		// steal from prev
		n.Keys = utils.Insert(n.Keys, 0, n.Prev.Keys[len(n.Prev.Keys)-1])
		n.Vals = utils.Insert(n.Vals, 0, n.Prev.Vals[len(n.Prev.Vals)-1])
		n.Prev.Keys = n.Prev.Keys[:len(n.Prev.Keys)-1]
		n.Prev.Vals = n.Prev.Vals[:len(n.Prev.Vals)-1]
	} else if n.Prev != nil && n.Prev.Parent == n.Parent {
		// merge with prev
		n.Prev.Keys = append(n.Prev.Keys, n.Keys...)
		n.Prev.Vals = append(n.Prev.Vals, n.Vals...)
		ourIdx, _ := n.Parent.findChildIdx(n)
		n.Parent.Keys = utils.Delete(n.Parent.Keys, ourIdx-1)
		n.Parent.Children = utils.Delete(n.Parent.Children, ourIdx)
		n.Prev.Next = n.Next
		if n.Next != nil {
			n.Next.Prev = n.Prev
		}
	} else {
		// merge with next
		n.Keys = append(n.Keys, n.Next.Keys...)
		n.Vals = append(n.Vals, n.Next.Vals...)
		ourIdx, _ := n.Parent.findChildIdx(n)
		n.Parent.Keys = utils.Delete(n.Parent.Keys, ourIdx)
		n.Parent.Children = utils.Delete(n.Parent.Children, ourIdx+1)
		if n.Next.Next != nil {
			n.Next.Next.Prev = n
		}
		n.Next = n.Next.Next
	}
	return n.Parent.merge()
}
