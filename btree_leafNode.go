package btree

import (
	"fmt"
	"strings"
)


type LeafNode[K Ordered, V any] struct {
	Order  int
	Keys   []K
	Vals   []V
	Parent *InternalNode[K, V]
	Next   *LeafNode[K, V]
	Prev   *LeafNode[K, V]
}

func NewLeafNode[K Ordered, V any](order int) *LeafNode[K, V] {
	return &LeafNode[K, V]{
		Order:    order,
		Keys:     make([]K, 0, order),
		Vals:     make([]V, 0, order),
		Parent:   nil,
		Next: nil,
		Prev: nil,
	}
}

func (n *LeafNode[K, V]) getParent() *InternalNode[K, V] {
	return n.Parent
}

func (n *LeafNode[K, V]) setParent(parent *InternalNode[K, V]) error {
	n.Parent = parent
	return nil
}

// returns the index of where the KVP would go
func (n *LeafNode[K, V]) find(key K) (int, bool) {
	for i, curKey := range n.Keys {
		if key == curKey {
			return i, true
		} else if key < curKey {
			return i, false
		}
	}
	return len(n.Keys), false
}

func (n *LeafNode[K, V]) get(key K) (V, bool) {
	// fmt.Printf("%v: %v %v\n", key, len(node.kvps), len(node.children))
	// fmt.Println(node.isLeaf)
	idx, wasFound := n.find(key)
	if wasFound {
		return n.Vals[idx], true
	}
	var zero V
	return zero, false
}

func (n *LeafNode[K, V]) getRange(start, end K, res []V) []V {
	startIdx, _ := n.find(start)
	endIdx, _ := n.find(end)
	res = append(res, n.Vals[startIdx:endIdx]...)
	if n.Next != nil && endIdx == len(n.Keys) {
		return n.Next.getRange(start, end, res)
	}
	return res
}

func (n *LeafNode[K, V]) traverse(res []V) []V {
	res = append(res, n.Vals...)
	if n.Next == nil {
		return res
	}
	return n.Next.traverse(res)
}

func (n *LeafNode[K, V]) set(key K, val V) (bool, error) {
	idx, wasFound := n.find(key)
	if wasFound {
		n.Vals[idx] = val
	} else {
		n.Keys = insert(n.Keys, idx, key)
		n.Vals = insert(n.Vals, idx, val)
	}
	return !wasFound, n.split()
}

func (n *LeafNode[K, V]) split() error {
	if len(n.Keys) < n.Order {
		return nil
	}

	siblingNode := NewLeafNode[K, V](n.Order)
	if n.Parent == nil {
		n.Parent = newInternalNode[K, V](n.Order)
		n.Parent.insertChild(0, n)
	}
	siblingNode.Parent = n.Parent

	mid := len(n.Keys) / 2

	idx, _ := n.Parent.find(n.Keys[mid])
	n.Parent.Keys = insert(n.Parent.Keys, idx, n.Keys[mid])
	n.Parent.insertChild(idx+1, siblingNode)

	siblingNode.Keys = make([]K, len(n.Keys[mid:]), n.Order)
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


func (n *LeafNode[K, V]) delete(key K) (bool, error) {
	idx, wasFound := n.find(key)
	if !wasFound {
		return false, nil
	}
	old := n.Keys[0]
	n.Keys = delete(n.Keys, idx)
	n.Vals = delete(n.Vals, idx)
	
	err := n.merge()
	if n.Parent != nil && len(n.Keys) > 0 {
		n.Parent.replaceKey(old, n.Keys[0])
	}
	return true, err
}

func (n *LeafNode[K, V]) merge() error {
	if n.Parent == nil || len(n.Keys) >= (n.Order-1)/2 {
		return nil
	}
	if n.Next != nil && n.Next.Parent == n.Parent &&
	len(n.Next.Keys) > (n.Order-1)/2 {
		// steal from next
		n.Parent.replaceKey(n.Next.Keys[0], n.Next.Keys[1])
		n.Keys = append(n.Keys, n.Next.Keys[0])
		n.Vals = append(n.Vals, n.Next.Vals[0])
		n.Next.Keys = n.Next.Keys[1:]
		n.Next.Vals = n.Next.Vals[1:]
	} else if n.Prev != nil && n.Prev.Parent == n.Parent &&
	len(n.Prev.Keys) > (n.Order-1)/2 {
		// steal from prev
		n.Keys = insert(n.Keys, 0, n.Prev.Keys[len(n.Prev.Keys)-1])
		n.Vals = insert(n.Vals, 0, n.Prev.Vals[len(n.Prev.Vals)-1])
		n.Prev.Keys = n.Prev.Keys[:len(n.Prev.Keys)-1]
		n.Prev.Vals = n.Prev.Vals[:len(n.Prev.Vals)-1]
	} else if n.Prev != nil && n.Prev.Parent == n.Parent {
		// merge with prev
		n.Prev.Keys = append(n.Prev.Keys, n.Keys...)
		n.Prev.Vals = append(n.Prev.Vals, n.Vals...)
		ourIdx, _ := n.Parent.findChildIdx(n)
		n.Parent.Keys = delete(n.Parent.Keys, ourIdx-1)
		n.Parent.Children = delete(n.Parent.Children, ourIdx)
		n.Prev.Next = n.Next
		if n.Next != nil {
			n.Next.Prev = n.Prev
		}
	} else {
		// merge with next
		n.Keys = append(n.Keys, n.Next.Keys...)
		n.Vals = append(n.Vals, n.Next.Vals...)
		ourIdx, _ := n.Parent.findChildIdx(n)
		n.Parent.Keys = delete(n.Parent.Keys, ourIdx)
		n.Parent.Children = delete(n.Parent.Children, ourIdx+1)
		if n.Next.Next != nil {
			n.Next.Next.Prev = n
		}
		n.Next = n.Next.Next
	}
	return n.Parent.merge()
}

func (n *LeafNode[K, V]) getNewRoot() BTreeNode[K, V] {
	if len(n.Keys) == 0 {
		return nil
	}
	return n
}

func (n *LeafNode[K, V]) print(level int) {
    indent := strings.Repeat("    ", level)
	next := n.Next
	prev := n.Prev
	nextKeys := []K{}
	if next != nil {
		nextKeys = next.Keys
	}
	prevKeys := []K{}
	if prev != nil {
		prevKeys = prev.Keys
	}
    fmt.Printf("%sLeaf Node: keys=%v, next=%v, prev=%v\n", indent, n.Keys, nextKeys, prevKeys)
}

func (n *LeafNode[K, V]) verify() (K, K) {
	return n.Keys[0], n.Keys[len(n.Keys)-1]
}

