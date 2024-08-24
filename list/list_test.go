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
