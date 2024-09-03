package btree

import (
	"log"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"
)

func intCmp(a, b int) int {
	return a - b
}

func TestRenderDot(t *testing.T) {
	t.Skip()

	bt := NewWithTee[int, string](intCmp, 4)
	bt.Insert(2, "two")
	bt.Insert(4, "four")
	bt.Insert(1, "one")
	bt.Insert(7, "seven")
	bt.Insert(6, "six")
	bt.Insert(3, "three")
	bt.Insert(11, "eleven")
	bt.Insert(9, "nine")
	bt.Insert(21, "twentyone")
	bt.Insert(8, "eight")
	bt.Insert(13, "thirteen")
	bt.Insert(14, "fourteen")
	bt.Insert(22, "22")
	bt.Insert(23, "23")
	bt.Insert(24, "24")
	bt.Insert(25, "21")
	//bt.renderDotToImage("bt.png")
	//fmt.Println(bt.findNodeForDeletion(bt.root, 21, nil))

	bt.Delete(1)
	bt.renderDotToImage("bt2.png")
}

func checkFound(t *testing.T, bt *BTree[int, string], key int, val string) {
	t.Helper()
	v, ok := bt.Get(key)
	if !ok {
		t.Errorf("key %v not found", key)
	}
	if v != val {
		t.Errorf("got %v, want %v", v, val)
	}
}

func checkNotFound(t *testing.T, bt *BTree[int, string], key int) {
	t.Helper()
	v, ok := bt.Get(key)
	if ok {
		t.Errorf("key %v found (%v)", key, v)
	}
}

func checkVerify[K, V any](t *testing.T, bt *BTree[K, V]) {
	if err := bt.verify(); err != nil {
		t.Error(err)
	}
}

func TestManualSmall(t *testing.T) {
	// Manually insert and get some nodes from a tree with t=4
	bt := NewWithTee[int, string](intCmp, 4)

	// Shouldn't find keys in empty tree
	checkNotFound(t, bt, 1)
	checkNotFound(t, bt, 4)
	checkNotFound(t, bt, 2)

	// Insert some keys
	bt.Insert(1, "1")
	bt.Insert(4, "4")

	// Find these, but not a key that wasn't inserted
	checkFound(t, bt, 1, "1")
	checkFound(t, bt, 4, "4")
	checkNotFound(t, bt, 2)

	// Insert some more keys to create a split (total > t*2-1)
	bt.Insert(2, "2")
	bt.Insert(3, "3")
	bt.Insert(11, "11")
	bt.Insert(8, "8")
	bt.Insert(5, "5")
	bt.Insert(9, "9")

	checkFound(t, bt, 3, "3")
	checkFound(t, bt, 8, "8")
	checkFound(t, bt, 1, "1")
	checkNotFound(t, bt, 6)

	// Override values
	bt.Insert(9, "99")
	bt.Insert(2, "22")

	checkFound(t, bt, 9, "99")
	checkFound(t, bt, 2, "22")

	checkVerify(t, bt)
}

func TestLargeSequential(t *testing.T) {
	// Insert a large number of nodes
	bt := NewWithTee[int, string](intCmp, 4)

	insertNumbersUpto(bt, 350)
	checkVerify(t, bt)

	for i := 1; i < 350; i++ {
		checkFound(t, bt, i, strconv.Itoa(i))
	}
}

func TestLargeStrings(t *testing.T) {
	s1 := rand.Uint64()
	s2 := rand.Uint64()
	log.Println("TestLargeSequential seed:", s1, s2)
	rnd := rand.New(rand.NewPCG(s1, s2))

	// New tree with default t
	bt := New[string, int](strings.Compare)

	// Create a map of string->serial number we'll be using for insertion and
	// later for comparison. The strings are not guaranteed unique, but that's ok,
	// since only the last instance will count for mp.
	N := 20000
	strs := make([]string, 0, N)
	for range N {
		strs = append(strs, randString(rnd, 5))
	}

	// mp keeps a stable mapping of string-->int
	// in the test we keep shuffling strs, but mp remains the same.
	mp := make(map[string]int)
	for i, s := range strs {
		mp[s] = i
	}

	// Insert all keys in some shuffled order.
	rand.Shuffle(len(strs), func(i, j int) {
		strs[i], strs[j] = strs[j], strs[i]
	})
	for i := 0; i < len(strs); i++ {
		bt.Insert(strs[i], mp[strs[i]])
	}
	checkVerify(t, bt)

	// Shuffle again and Get all strings in the shuffled order
	rand.Shuffle(len(strs), func(i, j int) {
		strs[i], strs[j] = strs[j], strs[i]
	})
	for i := 0; i < len(strs); i++ {
		v, ok := bt.Get(strs[i])
		if !ok || v != mp[strs[i]] {
			t.Errorf("not found or wrong value, got %v, want %v", v, mp[strs[i]])
		}
	}
}

func TestManualDeletion(t *testing.T) {
	bt := NewWithTee[int, string](intCmp, 4)

	for i := range 10 {
		bt.Insert(i*10, strconv.Itoa(i))
	}
	for i := range 10 {
		bt.Insert(i*10+1, strconv.Itoa(i))
		bt.Insert(i*10+2, strconv.Itoa(i))
	}

	checkVerify(t, bt)
	//bt.renderDotToImage("mn1.png")

	// Delete from leaf that has more than minimal: no rotation required
	bt.Delete(22)
	checkVerify(t, bt)

	// Delete from leaf that has a right sibling with enough elements to rotate
	// one key.
	bt.Delete(62)
	checkVerify(t, bt)

	// Delete from leaf that has a left sibling with enough elements to rotate
	// one key.
	bt.Delete(31)
	bt.Delete(32)
	checkVerify(t, bt)

	// Merge with right sibling
	bt.Delete(1)
	checkVerify(t, bt)

	// Merge with left sibling
	bt.renderDotToImage("mn1.png")
	bt.Delete(52)
	checkVerify(t, bt)
	bt.renderDotToImage("mn2.png")
}

// randString generates a random string made from lowercase chars with minimal
// length minLen; it uses rnd as the RNG state.
func randString(rnd *rand.Rand, minLen int) string {
	var rr []rune
	for i := 0; i < minLen; i++ {
		rr = append(rr, 'a'+rune(rnd.IntN(26)))
	}

	// Add more letters with p=0.8
	for rnd.IntN(5) <= 3 {
		rr = append(rr, 'a'+rune(rnd.IntN(26)))
	}
	return string(rr)
}

// insertNumbersUpto inserts numbers [1...upto] (inclusive) into bt into
// a predetermined shuffled order.
func insertNumbersUpto(bt *BTree[int, string], upto int) {
	rnd := rand.New(rand.NewPCG(9999, 404040))

	s := make([]int, upto)
	for i := 1; i <= upto; i++ {
		s[i-1] = i
	}
	rnd.Shuffle(len(s), func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})

	for _, num := range s {
		bt.Insert(num, strconv.Itoa(num))
	}
}
