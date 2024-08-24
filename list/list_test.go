package list

import (
	"slices"
	"testing"
)

func checkList[T comparable](t *testing.T, lst *List[T], want []T) {
	t.Helper()
	if lst.Len() != len(want) {
		t.Errorf("got len=%v, want %v", lst.Len(), len(want))
	}

	got := slices.Collect(lst.Values())
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if lst.front.prev != nil {
		t.Errorf("got lst.front.prev=%p, want nil", lst.front.prev)
	}
	if lst.back.next != nil {
		t.Errorf("got lst.back.next=%p, want nil", lst.back.next)
	}

	if lst.Len() < 1 {
		return
	}
	// Get all nodes from the list and verify list invariants.
	nodes := slices.Collect(lst.Nodes())
	first := nodes[0]
	last := nodes[len(nodes)-1]

	if lst.front.next != first || first.prev != lst.front {
		t.Errorf("front mismatch: front.next=%p, first.prev=%p", lst.front.next, first.prev)
	}
	if lst.back.prev != last || last.next != lst.back {
		t.Errorf("back mismatch: back.prev=%p, last.next=%p", lst.back.prev, last.next)
	}

	for i := 0; i < len(nodes)-1; i++ {
		j := i + 1
		if nodes[i].next != nodes[j] || nodes[j].prev != nodes[i] {
			t.Errorf("node link mismatch at i=%d, j=%d", i, j)
		}
	}
}

func TestBasicInsertFront(t *testing.T) {
	nl := New[int]()
	nl.InsertFront(20)
	checkList(t, nl, []int{20})
	nl.InsertFront(30)
	checkList(t, nl, []int{30, 20})
	nl.InsertFront(10)
	checkList(t, nl, []int{10, 30, 20})
}

func TestBasicInsertBack(t *testing.T) {
	nl := New[int]()
	nl.InsertBack(20)
	checkList(t, nl, []int{20})
	nl.InsertBack(30)
	checkList(t, nl, []int{20, 30})
	nl.InsertBack(10)
	checkList(t, nl, []int{20, 30, 10})
}

func TestInsertAfter(t *testing.T) {
	nl := New[int]()
	nl.InsertFront(99)
	nl.InsertAfter(nl.Front(), 102)
	checkList(t, nl, []int{99, 102})
	nl.InsertAfter(nl.Front(), 33)
	checkList(t, nl, []int{99, 33, 102})

	k := nl.InsertAfter(nl.Front(), 50)
	checkList(t, nl, []int{99, 50, 33, 102})
	nl.InsertAfter(k, 60)
	checkList(t, nl, []int{99, 50, 60, 33, 102})
}

func TestInsertBefore(t *testing.T) {
	nl := New[int]()
	nl.InsertFront(99)
	nl.InsertBefore(nl.Front(), 102)
	checkList(t, nl, []int{102, 99})
	nl.InsertBefore(nl.Front(), 33)
	checkList(t, nl, []int{33, 102, 99})

	k := nl.InsertBefore(nl.Back(), 50)
	checkList(t, nl, []int{33, 102, 50, 99})
	nl.InsertBefore(k, 60)
	checkList(t, nl, []int{33, 102, 60, 50, 99})
}

func TestRemove(t *testing.T) {
	nl := New[int]()
	nl.InsertBack(5)
	nl.InsertBack(6)
	nl.InsertBack(7)
	checkList(t, nl, []int{5, 6, 7})

	// Remove all elements from front
	nl.Remove(nl.Front())
	checkList(t, nl, []int{6, 7})
	nl.Remove(nl.Front())
	checkList(t, nl, []int{7})
	nl.Remove(nl.Front())
	checkList(t, nl, []int{})

	// Remove all elements from back
	nl = New[int]()
	nl.InsertBack(5)
	nl.InsertBack(6)
	nl.InsertBack(7)

	nl.Remove(nl.Back())
	checkList(t, nl, []int{5, 6})
	nl.Remove(nl.Back())
	checkList(t, nl, []int{5})
	nl.Remove(nl.Back())
	checkList(t, nl, []int{})
}

func TestFrontBack(t *testing.T) {
	// Empty list - nil for both
	nl := New[int]()
	if nl.Front() != nil {
		t.Errorf("got front=%v, want nil", nl.Front())
	}
	if nl.Back() != nil {
		t.Errorf("got back=%v, want nil", nl.Back())
	}

	// Insert element
	nl.InsertBack(50)
	if nl.Front().Value != 50 || nl.Back().Value != 50 {
		t.Errorf("got front=%v, back=%v, want 50", nl.Front().Value, nl.Back().Value)
	}
}

func TestNextPrev(t *testing.T) {
	nl := New[string]()
	nl.InsertBack("five")
	nl.InsertBack("six")
	nl.InsertBack("seven")

	var vals []string
	for n := nl.Front(); n != nil; n = nl.Next(n) {
		vals = append(vals, n.Value)
	}

	wantVals := []string{"five", "six", "seven"}
	if !slices.Equal(vals, wantVals) {
		t.Errorf("got %v, want %v", vals, wantVals)
	}

	var revVals []string
	for n := nl.Back(); n != nil; n = nl.Prev(n) {
		revVals = append(revVals, n.Value)
	}
	slices.Reverse(revVals)
	if !slices.Equal(revVals, wantVals) {
		t.Errorf("got %v, want %v", revVals, wantVals)
	}
}
