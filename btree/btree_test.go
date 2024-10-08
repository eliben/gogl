package btree

import (
	"log"
	"math/rand/v2"
	"slices"
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

func checkEmpty[K, V any](t *testing.T, bt *BTree[K, V]) {
	t.Helper()
	if !bt.root.leaf {
		t.Errorf("got leaf=true, want empty tree")
	}
	if len(bt.root.children) > 0 || len(bt.root.keys) > 0 {
		t.Errorf("got len(keys)=%d, len(children)=%d, want empty tree", len(bt.root.keys), len(bt.root.children))
	}
}

func checkVerify[K, V any](t *testing.T, bt *BTree[K, V]) {
	t.Helper()
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

	// Smoke test stats printing
	stats := bt.Stats()
	if strings.Index(stats, "Keys") < 0 {
		t.Errorf("got bad stats:\n%s\n", stats)
	}
}

// makeLoggedRand creates a new rand.Rand with a random source, and logs the
// source to output so the test can be reproduced if needed.
// To reproduce, create a rand.New(rand.NewPCG(s1, s2)) from the two seeds
// logged by this function.
func makeLoggedRand(t *testing.T) *rand.Rand {
	s1, s2 := rand.Uint64(), rand.Uint64()
	log.Printf("%s seed: %v, %v", t.Name(), s1, s2)
	return rand.New(rand.NewPCG(s1, s2))
}

func TestLargeSequential(t *testing.T) {
	rnd := makeLoggedRand(t)
	// Insert a large number of nodes
	bt := NewWithTee[int, string](intCmp, 4)
	h := newHarness(t, bt)
	insertNumbersUpto(rnd, h, 350)
	h.check()
}

func TestLargeStrings(t *testing.T) {
	rnd := makeLoggedRand(t)

	// New tree with default tee
	bt := New[string, int](strings.Compare)

	// Create a map of string->serial number we'll be using for insertion and
	// later for comparison. The strings are not guaranteed unique, but that's ok,
	// since only the last instance will count for mp.
	// We don't use the harness here because the types (just for this test) are
	// different.
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

// btHarness is a test harness for BTree[int, string] tests. It mirrors the
// insert/del operations into a map and for each one runs verification on the
// btree and ensures that all keys that are supposed to be in it are found
// successfully.
// Note that this means quadratic behavior for trees with many insertions.
// Use insertNoCheck and call check manually if needed.
type btHarness struct {
	t  *testing.T
	bt *BTree[int, string]
	m  map[int]string
}

func newHarness(t *testing.T, bt *BTree[int, string]) *btHarness {
	return &btHarness{
		t:  t,
		bt: bt,
		m:  make(map[int]string),
	}
}

func (bh *btHarness) insert(k int) {
	bh.insertNoCheck(k)
	bh.check()
}

func (bh *btHarness) insertNoCheck(k int) {
	v := strconv.Itoa(k)
	bh.bt.Insert(k, v)
	bh.m[k] = v
}

func (bh *btHarness) del(k int) {
	bh.bt.Delete(k)
	delete(bh.m, k)

	bh.check()
}

func (bh *btHarness) check() {
	checkVerify(bh.t, bh.bt)
	for k, v := range bh.m {
		checkFound(bh.t, bh.bt, k, v)
	}
}

func TestManualDeletionLeavesOnly(t *testing.T) {
	bt := NewWithTee[int, string](intCmp, 4)
	h := newHarness(t, bt)

	for i := range 10 {
		h.insert(i * 10)
	}
	for i := range 10 {
		h.insert(i*10 + 1)
		h.insert(i*10 + 2)
	}

	// Delete from leaf that has more than minimal: no rotation required
	h.del(22)

	// Delete from leaf that has a right sibling with enough elements to rotate
	// one key.
	h.del(62)

	// Delete from leaf that has a left sibling with enough elements to rotate
	// one key.
	h.del(31)
	h.del(32)

	// Merge with right sibling
	h.del(1)

	// Merge with left sibling
	h.del(52)

	// These should be safe no-ops
	for i := 100; i < 200; i++ {
		h.del(i)
	}
}

func TestDeleteAllSmall(t *testing.T) {
	bt := NewWithTee[int, string](intCmp, 4)
	h := newHarness(t, bt)
	h.insert(1)
	h.insert(4)
	h.insert(9)
	h.insert(2)

	h.del(1)
	h.del(2)
	h.del(9)
	h.del(4)
	checkEmpty(t, bt)

	h.insert(3)
	h.insert(7)
	checkNotFound(t, bt, 1)
	checkNotFound(t, bt, 2)
}

func TestDeleteAllLarge(t *testing.T) {
	rnd := makeLoggedRand(t)

	// This test is repeated:
	//
	// * Insert some random elements into the tree (their number is also
	//   random).
	// * Remove all elements from the tree, one by one. Removal with
	//   btHarness.del runs .check on the tree after every deletion.
	// * Check that the tree is properly empty at the end of the process.
	for tn := range 10 {
		t.Run(strconv.Itoa(tn), func(t *testing.T) {
			bt := NewWithTee[int, string](intCmp, 4)
			h := newHarness(t, bt)

			numElems := 200 + rnd.IntN(400)
			rs := randomIntSlice(rnd, numElems, 10000)
			for _, n := range rs {
				h.insertNoCheck(n)
			}

			// Reshuffle rs to delete in an arbitrary order.
			rnd.Shuffle(len(rs), func(i, j int) {
				rs[i], rs[j] = rs[j], rs[i]
			})
			for _, n := range rs {
				h.del(n)
			}
			checkEmpty(t, bt)
		})
	}
}

func TestChurn(t *testing.T) {
	rnd := makeLoggedRand(t)

	// This test is repeated:
	//
	// * Insert some random elements into the tree.
	// * In a loop:
	//   * Insert some new elements
	//   * Delete some elements
	//   * (both of the last steps perform tree verification after each step)
	for tn := range 10 {
		t.Run(strconv.Itoa(tn), func(t *testing.T) {
			bt := NewWithTee[int, string](intCmp, 4)
			h := newHarness(t, bt)

			numElems := tn*50 + rnd.IntN(40)
			rs := randomIntSlice(rnd, numElems, 10000)
			for _, n := range rs {
				h.insertNoCheck(n)
			}

			for range 100 {
				numInserts := 2 + rnd.IntN(1+tn*2)
				for range numInserts {
					elem := rnd.IntN(10000)
					rs = append(rs, elem)
					h.insert(elem)
				}

				numDels := 2 + rnd.IntN(1+tn*2)
				for range numDels {
					elem := rs[rnd.IntN(len(rs))]
					h.del(elem)
					rs = slices.DeleteFunc(rs, func(e int) bool { return e == elem })
					if len(rs) == 0 {
						break
					}
				}
			}
		})
	}
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

// insertNumbersUpto inserts numbers [1...upto] (inclusive) into a harness into
// a random shuffled order.
func insertNumbersUpto(rnd *rand.Rand, h *btHarness, upto int) {
	s := make([]int, upto)
	for i := 1; i <= upto; i++ {
		s[i-1] = i
	}
	rnd.Shuffle(len(s), func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})

	for _, num := range s {
		h.insertNoCheck(num)
	}
}

// randomIntSlice creates a slice of n random integers in the range [0,upto).
// This function may run slowly unless n << upto.
func randomIntSlice(rnd *rand.Rand, n int, upto int) []int {
	var result []int
	// Create a set to ensure we get n non-duplicate ints.
	m := make(map[int]bool)
	for len(result) < n {
		num := rnd.IntN(upto)
		if !m[num] {
			m[num] = true
			result = append(result, num)
		}
	}
	return result
}
