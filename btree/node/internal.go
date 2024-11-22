package node

import (
	"btree/utils"
	"fmt"
	"strings"
)

func newInternalNode(order uint32) *InternalNode {
	return &InternalNode{
		Order:    order,
		Keys:     make([]int32, 0, order),
		Children: make([]BTreeNode, 0, order+1),
		Parent:   nil,
	}
}

func (n *InternalNode) insertChild(idx int, child BTreeNode) error {
	n.Children = utils.Insert(n.Children, idx, child)
	return nil
}

func (n *InternalNode) getParent() *InternalNode {
	return n.Parent
}

func (n *InternalNode) setParent(parent *InternalNode) error {
	n.Parent = parent
	return nil
}

func (n *InternalNode) find(key int32) (int, bool) {
	for i, curKey := range n.Keys {
		if key < curKey {
			return i, true
		}
	}
	return len(n.Keys), false
}

func (n *InternalNode) findChildIdx(child BTreeNode) (int, bool) {
	for i, curChild := range n.Children {
		if curChild == child {
			return i, true
		}
	}
	return -1, false
}

func (n *InternalNode) get(key int32) ([]byte, bool) {
	idx, _ := n.find(key)
	return n.Children[idx].get(key)
}

func (n *InternalNode) getRange(start, end int32, res [][]byte) [][]byte {
	idx, _ := n.find(start)
	return n.Children[idx].getRange(start, end, res)
}

func (n *InternalNode) traverse(res [][]byte) [][]byte {
	return n.Children[0].traverse(res)
}

func (n *InternalNode) set(key int32, val []byte) (bool, error) {
	idx, _ := n.find(key)
	return n.Children[idx].set(key, val)
}

func (n *InternalNode) delete(key int32) (bool, error) {
	idx, _ := n.find(key)
	return n.Children[idx].delete(key)
}

func (n *InternalNode) replaceKey(oldKey int32, newKey int32) {
	if oldKey == newKey {
		return
	} else if len(n.Keys) == 0 || oldKey < n.Keys[0] {
		if n.Parent != nil {
			n.Parent.replaceKey(oldKey, newKey)
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

func (n *InternalNode) split() error {
	if len(n.Keys) < int(n.Order) {
		return nil
	}

	siblingNode := newInternalNode(n.Order)
	if n.Parent == nil {
		n.Parent = newInternalNode(n.Order)
		n.Parent.insertChild(0, n)
	}
	siblingNode.Parent = n.Parent

	mid := len(n.Keys) / 2

	idx, _ := n.Parent.find(n.Keys[mid])
	n.Parent.Keys = utils.Insert(n.Parent.Keys, idx, n.Keys[mid])
	n.Parent.insertChild(idx+1, siblingNode)

	siblingNode.Keys = make([]int32, len(n.Keys[mid+1:]), n.Order)
	copy(siblingNode.Keys, n.Keys[mid+1:])
	n.Keys = n.Keys[:mid]

	siblingNode.Children = make([]BTreeNode, len(n.Children[mid+1:]), n.Order+1)
	copy(siblingNode.Children, n.Children[mid+1:])
	n.Children = n.Children[:mid+1]

	for _, child := range siblingNode.Children {
		child.setParent(siblingNode)
	}

	return n.Parent.split()
}

func (n *InternalNode) merge() error {
	if n.Parent == nil || len(n.Children) >= int(n.Order+1)/2 {
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
		len(n.Parent.Children[ourIdx-1].(*InternalNode).Children) > int(n.Order+1)/2 {
		// Steal from left sibling
		leftSibling := n.Parent.Children[ourIdx-1].(*InternalNode)
		n.Keys = utils.Insert(n.Keys, 0, n.Parent.Keys[ourIdx-1])
		n.Parent.Keys[ourIdx-1] = leftSibling.Keys[len(leftSibling.Keys)-1]
		leftSibling.Keys = leftSibling.Keys[:len(leftSibling.Keys)-1]
		n.Children = utils.Insert(n.Children, 0, leftSibling.Children[len(leftSibling.Children)-1])
		leftSibling.Children = leftSibling.Children[:len(leftSibling.Children)-1]
		n.Children[0].setParent(n)

		return nil
	} else if ourIdx < len(n.Parent.Children)-1 &&
		len(n.Parent.Children[ourIdx+1].(*InternalNode).Children) > int(n.Order+1)/2 {
		// Steal from right sibling
		rightSibling := n.Parent.Children[ourIdx+1].(*InternalNode)
		n.Keys = append(n.Keys, n.Parent.Keys[ourIdx])
		n.Parent.Keys[ourIdx] = rightSibling.Keys[0]
		rightSibling.Keys = rightSibling.Keys[1:]
		rightSibling.Children = rightSibling.Children[1:]
		n.Children[len(n.Children)-1].setParent(n)

		return nil
	}

	if ourIdx > 0 {
		// Merge with left sibling
		leftSibling := n.Parent.Children[ourIdx-1].(*InternalNode)
		leftSibling.Keys = append(leftSibling.Keys, n.Parent.Keys[ourIdx-1])
		leftSibling.Keys = append(leftSibling.Keys, n.Keys...)
		leftSibling.Children = append(leftSibling.Children, n.Children...)
		for _, child := range n.Children {
			child.setParent(leftSibling)
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
			child.setParent(n)
		}
		n.Parent.Keys = utils.Delete(n.Parent.Keys, ourIdx)
		n.Parent.Children = utils.Delete(n.Parent.Children, ourIdx+1)

		return n.Parent.merge()
	}
}

func (n *InternalNode) getNewRoot() BTreeNode {
	if len(n.Keys) == 0 {
		return n.Children[0]
	}
	return n
}

func (n *InternalNode) print(level int) {
	indent := strings.Repeat("    ", level)
	fmt.Printf("%sInternal Node: keys=%v\n", indent, n.Keys)
	for _, child := range n.Children {
		child.print(level + 1)
	}
}

func (n *InternalNode) verify() (int32, int32) {
	for i := range n.Keys {
		_, wasFound := n.Children[i+1].get(n.Keys[i])
		if !wasFound {
			panic(fmt.Sprintf("key %v not found in node %v\n", n.Keys[i], n))
		}
	}
	a, _ := n.Children[0].verify()
	_, b := n.Children[len(n.Children)-1].verify()
	for i := 0; i < len(n.Keys); i++ {
		x, _ := n.Children[i].verify()
		if x >= n.Keys[i] {
			panic("uh oh\n")
		}
		y, _ := n.Children[i+1].verify()
		if y != n.Keys[i] {
			panic("uh oh\n")
		}
	}
	for _, child := range n.Children {
		child.verify()
	}
	return a, b
}
