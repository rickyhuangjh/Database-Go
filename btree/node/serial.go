package node

import (
	"fmt"
	"bytes"
	"encoding/binary"
)


func Serialize(n BTreeNode) []byte {
	switch n.(type) {
	case *InternalNode:
		return serializeInternal(n.(*InternalNode))
	default:
		return serializeLeaf(n.(*LeafNode))
	}
}

func Deserialize(b []byte) BTreeNode {
	if uint8(b[7]) & (1 << 7) == 0 {
		return deserializeInternal(b)
	} else {
		return deserializeLeaf(b)
	}
}

func serializeInternal(n *InternalNode) []byte {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, n.ID)
	if err != nil {
		panic("Could not serialize Internal ID")
	}
	err = binary.Write(buf, binary.LittleEndian, n.ParentID)
	if err != nil {
		panic("Could not serialize Internal ParentID")
	}
	err = binary.Write(buf, binary.LittleEndian, uint64(len(n.ChildIDs)))
	if err != nil {
		panic("Could not serialize Internal length")
	}
	var key uint64
	for i, childID := range(n.ChildIDs) {
		if i < len(n.Keys) {
			key = n.Keys[i]
		} else {
			key = 0
		}
		err = binary.Write(buf, binary.LittleEndian, key)
		if err != nil {
			fmt.Println(err)
			panic("Could not serialize Internal Key")
		}
		err = binary.Write(buf, binary.LittleEndian, childID)
		if err != nil {
			panic("Could not serialize Internal childID")
		}
	}
	return buf.Bytes()
}

func serializeLeaf(n *LeafNode) []byte {
	buf := new(bytes.Buffer)
	
	err := binary.Write(buf, binary.LittleEndian, n.ID)
	if err != nil {
		panic("Could not serialize Internal ID")
	}
	err = binary.Write(buf, binary.LittleEndian, n.ParentID)
	if err != nil {
		panic("Could not serialize Internal ParentID")
	}
	err = binary.Write(buf, binary.LittleEndian, n.NextID)
	if err != nil {
		panic("Could not serialize Internal NextID")
	}
	err = binary.Write(buf, binary.LittleEndian, n.PrevID)
	if err != nil {
		panic("Could not serialize Internal PrevID")
	}
	err = binary.Write(buf, binary.LittleEndian, uint64(len(n.Vals)))
	if err != nil {
		fmt.Println(err)
		panic("Could not serialize Internal length")
	}
	for i, val := range(n.Vals) {
		key := n.Keys[i]
		err = binary.Write(buf, binary.LittleEndian, key)
		if err != nil {
			panic("Could not serialize Internal key")
		}
		err = binary.Write(buf, binary.LittleEndian, val)
		if err != nil {
			panic("Could not serialize Internal childID")
		}
	}
	return buf.Bytes()
}

func deserializeInternal(b []byte) *InternalNode {
	node := NewInternalNode()
	reader := bytes.NewReader(b)
	
	err := binary.Read(reader, binary.LittleEndian, &node.ID)
	if err != nil {
		panic("Could not deserialize Internal ID")
	}
	err = binary.Read(reader, binary.LittleEndian, &node.ParentID)
	if err != nil {
		panic("Could not deserialize Internal ParentID")
	}
	var length uint64
	err = binary.Read(reader, binary.LittleEndian, &length)
	if err != nil {
		panic("Could not deserialize Internal length")
	}
	var key uint64
	var childID uint64
	for i:=uint64(0); i<length; i++ {
		err = binary.Read(reader, binary.LittleEndian, &key)
		if err != nil {
			fmt.Println(err)
			panic("Coud not deserialize Internal Key")
		}
		err = binary.Read(reader, binary.LittleEndian, &childID)
		if err != nil {
			panic("Coud not deserialize Internal ChildID")
		}
		if i < length - 1 {
			node.Keys = append(node.Keys, key)
		}
		node.ChildIDs = append(node.ChildIDs, childID)
	}
	return node
}


func deserializeLeaf(b []byte) *LeafNode {
	node := NewLeafNode()
	reader := bytes.NewReader(b)
	
	err := binary.Read(reader, binary.LittleEndian, &node.ID)
	if err != nil {
		panic("Could not deserialize Leaf ID")
	}
	err = binary.Read(reader, binary.LittleEndian, &node.ParentID)
	if err != nil {
		panic("Could not deserialize Leaf ParentID")
	}
	err = binary.Read(reader, binary.LittleEndian, &node.NextID)
	if err != nil {
		panic("Could not deserialize Leaf NextID")
	}
	err = binary.Read(reader, binary.LittleEndian, &node.PrevID)
	if err != nil {
		panic("Could not deserialize Leaf PrevID")
	}
	var length uint64
	err = binary.Read(reader, binary.LittleEndian, &length)
	if err != nil {
		panic("Could not deserialize Leaf length")
	}
	var key uint64
	var val uint64
	for i:=uint64(0); i<length; i++ {
		err = binary.Read(reader, binary.LittleEndian, &key)
		if err != nil {
			fmt.Println(err)
			panic("Coud not deserialize Leaf Key")
		}
		err = binary.Read(reader, binary.LittleEndian, &val)
		if err != nil {
			panic("Coud not deserialize Leaf Val")
		}
		node.Keys = append(node.Keys, key)
		node.Vals = append(node.Vals, val)
	}
	return node
}

