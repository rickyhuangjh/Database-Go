package btree


import (
	"encoding/gob"
	"bytes"
	"fmt"
)


type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
	~float32 | ~float64 | ~string
}

type BTree[K Ordered, V any] interface {
	Set(key K, val V) error
	Delete(key K) error
	Get(key K) (V, bool)
	GetRange(start, end K) []V
	Traverse() []V
	Print()
	Verify()
	Serialize() ([]byte, error)
}

type btree[K Ordered, V any] struct {
	Order int
	Root  BTreeNode[K, V]
	Size  int
}

func NewBTree[K Ordered, V any](order int) BTree[K, V] {
	return &btree[K, V]{
		Order: order,
		Root:  nil,
		Size:  0,
	}
}

func (b *btree[K, V]) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	gob.Register(LeafNode[K, V]{})
	gob.Register(InternalNode[K, V]{})
	enc := gob.NewEncoder(buf)
	err := enc.Encode(b)
	return buf.Bytes(), err
}

func Deserialize[K Ordered, V any](data []byte) (BTree[K, V], error) {
	b := new(btree[K, V])
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(b)
	return b, err
}


func (b *btree[K, V]) Get(key K) (V, bool) {
	return b.Root.get(key)
}

func (b *btree[K, V]) Traverse() []V {
	if b.Root == nil {
		return nil
	}
	res := b.Root.traverse(make([]V, 0, b.Size))
	return res
}

func (b *btree[K, V]) Set(key K, val V) error {
	if b.Root == nil {
		b.Root = NewLeafNode[K, V](b.Order)
	}

	wasAdded, err := b.Root.set(key, val)
	if b.Root.getParent() != nil {
		// fmt.Println("hello")
		b.Root = b.Root.getParent()
	}
	if wasAdded {
		b.Size++
	}
	//fmt.Println(a)
	return err
}

func (b *btree[K, V]) Delete(key K) error {
	if b.Root == nil {
		return nil
	}
	wasDeleted, err := b.Root.delete(key)
	if wasDeleted {
			b.Size--
	}
	b.Root = b.Root.getNewRoot()
	if b.Root != nil {
		b.Root.setParent(nil)
	}
	return err
}

func (b *btree[K, V]) Print() {
    if b.Root == nil {
        fmt.Println("Empty tree")
        return
    }
    b.Root.print(0)
}

func (b *btree[K, V]) Verify() {
	if b.Root == nil {
		return
	}
	b.Root.verify()
}

func (b *btree[K, V]) GetRange(start, end K) []V {
	res := make([]V, 0)
	res = b.Root.getRange(start, end, res)
	return res
}
