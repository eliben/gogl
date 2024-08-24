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

	got := slices.Collect(lst.All())
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
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
