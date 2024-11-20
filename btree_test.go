package btree

import (
    "testing"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const items = 1e6

func TestSet(t *testing.T) {
	btree := NewBTree[int, int](7)
	for i:=0; i<items; i++ {
		num := rand.Intn(items)
		btree.Set(num, num)
		if i % (items/20) == 0 {
			fmt.Printf("%v\n", float64(i)/float64(items))
		}
	}
	data, err := btree.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile("btree.dat", data, 0644)
}

func TestGet(t *testing.T) {
	data, err := os.ReadFile("btree.dat")
	if err != nil {
		t.Fatal(err)
	}
	btree, err := Deserialize[int, int](data)
	if err != nil {
		t.Fatal(err)
	}
	btree.Print()
}

func TestInsert(t *testing.T) {
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"
	
	strings := make([]string, items)
	
	for i:=0; i<items; i++ {
		length := rand.Intn(10) + 1
		for j:=0; j<length; j++ {
			strings[i] += string(charset[rand.Intn(len(charset))])
		}
	}
	btree := NewBTree[string, string](7)
	m := make(map[string]string)
	start := time.Now()
	for i:=0; i<items; i++ {
		btree.Set(strings[i], strings[i])
	}
	elapsed := time.Since(start)
	fmt.Printf("btree insert time: %s\n", elapsed)
	start = time.Now()
	for i:=0; i<items; i++ {
		m[strings[i]] = strings[i]
	}
	elapsed = time.Since(start)
	fmt.Printf("map insert time: %s\n", elapsed)
}


