package btree

import (
	"fmt"
	"strings"
)

func insert[T any](slice []T, idx int, val T) []T {
	if idx < 0 || idx > len(slice) {
		panic("Slice insert idx out of bounds")
	}
	if idx == cap(slice) {
		panic("Slice over capacity")
	}
	var dummy T
	slice = append(slice, dummy)
	copy(slice[idx+1:], slice[idx:])
	slice[idx] = val
	return slice
}

func delete[T any](slice []T, idx int) []T {
	if idx < 0 || idx >= len(slice) {
		panic("Slice delete idx out of bounds")
	}
	return append(slice[:idx], slice[idx+1:]...)
}


type BTreeNode[K Ordered, V any] interface {
	get(key K) (V, bool)
	getRange(start, end K, res []V) []V
	traverse(res []V) []V
	set(key K, val V) (bool, error)
	delete(key K) (bool, error)
	getParent() *InternalNode[K, V]
	setParent(parent *InternalNode[K, V]) error
	split() error
	getNewRoot() BTreeNode[K, V]
	print(level int)
	verify() (K, K)
}

type InternalNode[K Ordered, V any] struct {
	Order    int
	Keys     []K
	Children []BTreeNode[K, V]
	Parent   *InternalNode[K, V]
}

func newInternalNode[K Ordered, V any](order int) *InternalNode[K, V] {
	return &InternalNode[K, V]{
		Order:    order,
		Keys:     make([]K, 0, order),
		Children: make([]BTreeNode[K, V], 0, order+1),
		Parent:   nil,
	}
}

func (node *InternalNode[K, V]) insertChild(idx int, child BTreeNode[K, V]) error {
	node.Children = insert(node.Children, idx, child)
	return nil
}

func (node *InternalNode[K, V]) getParent() *InternalNode[K, V] {
	return node.Parent
}
func (node *InternalNode[K, V]) setParent(parent *InternalNode[K, V]) error {
	node.Parent = parent
	return nil
}

// returns the index of where the child would go
func (node *InternalNode[K, V]) find(key K) (int, bool) {
	for i, curKey := range node.Keys {
		if key < curKey {
			return i, true
		}
	}
	return len(node.Keys), false
}

func (node *InternalNode[K, V]) findChildIdx(child BTreeNode[K, V]) (int, bool) {
	for i, curChild := range node.Children {
		if curChild == child {
			return i, true
		}
	}
	return -1, false
}

func (node *InternalNode[K, V]) get(key K) (V, bool) {
	idx, _ := node.find(key)
	return node.Children[idx].get(key)
}

func (node *InternalNode[K, V]) getRange(start, end K, res []V) []V {
	idx, _ := node.find(start)
	return node.Children[idx].getRange(start, end, res)
}

func (node *InternalNode[K, V]) traverse(res []V) []V {
	return node.Children[0].traverse(res)
}

func (node *InternalNode[K, V]) set(key K, val V) (bool, error) {
	idx, _ := node.find(key)
	return node.Children[idx].set(key, val)
}

func (node *InternalNode[K, V]) delete(key K) (bool, error) {
	idx, _ := node.find(key)
	return node.Children[idx].delete(key)
}

func (node *InternalNode[K, V]) replaceKey(oldKey K, newKey K) {
	if oldKey == newKey {
		return
	} else if len(node.Keys) == 0 || oldKey < node.Keys[0] {
		if node.Parent != nil {
			node.Parent.replaceKey(oldKey, newKey)
		}
		return
	}
	for i, curKey := range node.Keys {
		if oldKey == curKey{
			node.Keys[i] = newKey
			return
		}
	}
}

func (node *InternalNode[K, V]) split() error {
	if len(node.Keys) < node.Order {
		return nil
	}

	siblingNode := newInternalNode[K, V](node.Order)
	if node.Parent == nil {
		node.Parent = newInternalNode[K, V](node.Order)
		node.Parent.insertChild(0, node)
	}
	siblingNode.Parent = node.Parent

	mid := len(node.Keys) / 2

	idx, _ := node.Parent.find(node.Keys[mid])
	node.Parent.Keys = insert(node.Parent.Keys, idx, node.Keys[mid])
	node.Parent.insertChild(idx+1, siblingNode)

	siblingNode.Keys = make([]K, len(node.Keys[mid+1:]), node.Order)
	copy(siblingNode.Keys, node.Keys[mid+1:])
	node.Keys = node.Keys[:mid]

	siblingNode.Children = make([]BTreeNode[K, V], len(node.Children[mid+1:]), node.Order+1)
	copy(siblingNode.Children, node.Children[mid+1:])
	node.Children = node.Children[:mid+1]

	for _, child := range siblingNode.Children {
		child.setParent(siblingNode)
	}

	return node.Parent.split()
}

func (node *InternalNode[K, V]) merge() error {
    if node.Parent == nil ||len(node.Keys) >= (node.Order+1)/2 - 1 {
        return nil
    }

    ourIdx := 0
    for i, child := range node.Parent.Children {
        if child == node {
            ourIdx = i
            break
        }
    }

    // Try to steal from siblings first
    if ourIdx > 0 && 
		len(node.Parent.Children[ourIdx-1].(*InternalNode[K, V]).Keys) > (node.Order+1)/2-1 {
        // Steal from left sibling
        leftSibling := node.Parent.Children[ourIdx-1].(*InternalNode[K, V])
        node.Keys = insert(node.Keys, 0, node.Parent.Keys[ourIdx-1])
        node.Parent.Keys[ourIdx-1] = leftSibling.Keys[len(leftSibling.Keys)-1]
        leftSibling.Keys = leftSibling.Keys[:len(leftSibling.Keys)-1]
        node.Children = insert(node.Children, 0, leftSibling.Children[len(leftSibling.Children)-1])
        leftSibling.Children = leftSibling.Children[:len(leftSibling.Children)-1]
        node.Children[0].setParent(node)
        
        return nil
    } else if ourIdx < len(node.Parent.Children)-1 &&
		len(node.Parent.Children[ourIdx+1].(*InternalNode[K, V]).Keys) > (node.Order+1)/2-1 {
        // Steal from right sibling
        rightSibling := node.Parent.Children[ourIdx+1].(*InternalNode[K, V])
        node.Keys = append(node.Keys, node.Parent.Keys[ourIdx])
        node.Parent.Keys[ourIdx] = rightSibling.Keys[0]
        rightSibling.Keys = rightSibling.Keys[1:]
        rightSibling.Children = rightSibling.Children[1:]
        node.Children[len(node.Children)-1].setParent(node)
        
        return nil
    }

    if ourIdx > 0 {
        // Merge with left sibling
        leftSibling := node.Parent.Children[ourIdx-1].(*InternalNode[K, V])
        leftSibling.Keys = append(leftSibling.Keys, node.Parent.Keys[ourIdx-1])
        leftSibling.Keys = append(leftSibling.Keys, node.Keys...)
        leftSibling.Children = append(leftSibling.Children, node.Children...)
        for _, child := range node.Children {
            child.setParent(leftSibling)
        }
        node.Parent.Keys = delete(node.Parent.Keys, ourIdx-1)
        node.Parent.Children = delete(node.Parent.Children, ourIdx)
        
        return node.Parent.merge()
    } else {
        // Merge with right sibling
        rightSibling := node.Parent.Children[ourIdx+1].(*InternalNode[K, V])
        node.Keys = append(node.Keys, node.Parent.Keys[ourIdx])
        node.Keys = append(node.Keys, rightSibling.Keys...)
        node.Children = append(node.Children, rightSibling.Children...)
        for _, child := range rightSibling.Children {
            child.setParent(node)
        }
        node.Parent.Keys = delete(node.Parent.Keys, ourIdx)
        node.Parent.Children = delete(node.Parent.Children, ourIdx+1)
        
        return node.Parent.merge()
    }
}

func (node *InternalNode[K, V]) getNewRoot() BTreeNode[K, V] {
	if len(node.Keys) == 0 {
		return node.Children[0]
	}
	return node
}

func (node *InternalNode[K, V]) print(level int) {
    indent := strings.Repeat("    ", level)
    fmt.Printf("%sInternal Node: keys=%v\n", indent, node.Keys)
    for _, child := range node.Children {
        child.print(level + 1)
    }
}

func (node *InternalNode[K, V]) verify() (K, K) {
	for i := range node.Keys {
		_, wasFound := node.Children[i+1].get(node.Keys[i])
		if !wasFound {
			panic(fmt.Sprintf("key %v not found in node %v\n", node.Keys[i], node))
		}
	}
	a, _ := node.Children[0].verify()
	_, b := node.Children[len(node.Children)-1].verify()
	for i:=0; i<len(node.Keys); i++ {
		x, _ := node.Children[i].verify()
		if x >= node.Keys[i] {
			panic("uh oh\n")
		}
		y, _ := node.Children[i+1].verify()
		if y != node.Keys[i] {
			panic("uh oh\n")
		}
	}
	for _, child := range node.Children {
		child.verify()
	}
	return a, b
}
