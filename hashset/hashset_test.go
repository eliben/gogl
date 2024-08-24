package hashset

import (
	"slices"
	"testing"
)

func TestAll(t *testing.T) {
	hs := New[int]()

	checkAll := func(wantSorted []int) {
		t.Helper()
		if hs.Len() != len(wantSorted) {
			t.Errorf("got len=%v, want %v", hs.Len(), len(wantSorted))
		}

		got := slices.Sorted(hs.All())
		if !slices.Equal(got, wantSorted) {
			t.Errorf("got %v, want %v", got, wantSorted)
		}
	}
	checkAll([]int{})
	hs.Add(10)
	checkAll([]int{10})

	hs.Add(20)
	hs.Add(13)
	checkAll([]int{10, 13, 20})

	hs.Add(18)
	checkAll([]int{10, 13, 18, 20})

	hs.Delete(18)
	checkAll([]int{10, 13, 20})
	hs.Delete(10)
	checkAll([]int{13, 20})

	hs.Add(50)
	hs.Add(5)
	checkAll([]int{5, 13, 20, 50})

	hs.Add(60)
	hs.Add(60)
	hs.Add(60)
	checkAll([]int{5, 13, 20, 50, 60})

	hs.Delete(60)
	hs.Delete(60)
	hs.Delete(60)
	checkAll([]int{5, 13, 20, 50})

	hs.Delete(50)
	hs.Delete(20)
	hs.Delete(5)
	checkAll([]int{13})

	hs.Delete(13)
	checkAll([]int{})
}

func TestContains(t *testing.T) {
	hs := New[string]()

	checkContains := func(v string, want bool) {
		t.Helper()
		got := hs.Contains(v)
		if got != want {
			t.Errorf("contains(%v)=%v, want %v", v, got, want)
		}
	}

	checkContains("joe", false)
	hs.Add("joe")
	checkContains("joe", true)
	hs.Delete("joe")
	checkContains("joe", false)

	hs.Add("bee")
	hs.Add("geranium")
	checkContains("joe", false)
	checkContains("bee", true)
	checkContains("geranium", true)

	hs.Add("cheese")
	hs.Add("io")
	hs.Add("joe")

	for _, v := range []string{"joe", "bee", "geranium", "io", "cheese"} {
		checkContains(v, true)
		hs.Delete(v)
	}

	for _, v := range []string{"joe", "bee", "geranium", "io", "cheese"} {
		checkContains(v, false)
	}
}
