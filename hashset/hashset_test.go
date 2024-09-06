package hashset

import (
	"slices"
	"testing"
)

func checkAll(t *testing.T, hs *HashSet[int], wantSorted []int) {
	t.Helper()
	if hs.Len() != len(wantSorted) {
		t.Errorf("got len=%v, want %v", hs.Len(), len(wantSorted))
	}

	got := slices.Sorted(hs.All())
	if !slices.Equal(got, wantSorted) {
		t.Errorf("got %v, want %v", got, wantSorted)
	}
}

func TestAll(t *testing.T) {
	hs := New[int]()

	checkAll(t, hs, []int{})
	hs.Add(10)
	checkAll(t, hs, []int{10})

	hs.Add(20)
	hs.Add(13)
	checkAll(t, hs, []int{10, 13, 20})

	hs.Add(18)
	checkAll(t, hs, []int{10, 13, 18, 20})

	hs.Delete(18)
	checkAll(t, hs, []int{10, 13, 20})
	hs.Delete(10)
	checkAll(t, hs, []int{13, 20})

	hs.Add(50)
	hs.Add(5)
	checkAll(t, hs, []int{5, 13, 20, 50})

	hs.Add(60)
	hs.Add(60)
	hs.Add(60)
	checkAll(t, hs, []int{5, 13, 20, 50, 60})

	hs.Delete(60)
	hs.Delete(60)
	hs.Delete(60)
	checkAll(t, hs, []int{5, 13, 20, 50})

	hs.Delete(50)
	hs.Delete(20)
	hs.Delete(5)
	checkAll(t, hs, []int{13})

	hs.Delete(13)
	checkAll(t, hs, []int{})
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

func TestSetOperations(t *testing.T) {
	hs1 := InitWith(10, 20, 30, 40)

	hs11 := InitWith(11, 21, 30, 41)
	u1 := hs1.Union(hs11)
	checkAll(t, u1, []int{10, 11, 20, 21, 30, 40, 41})
	i1 := hs1.Intersection(hs11)
	checkAll(t, i1, []int{30})
	d1 := hs1.Difference(hs11)
	checkAll(t, d1, []int{10, 20, 40})

	hs22 := InitWith(20)
	u2 := hs1.Union(hs22)
	checkAll(t, u2, []int{10, 20, 30, 40})
	i2 := hs1.Intersection(hs22)
	checkAll(t, i2, []int{20})
	d2 := hs1.Difference(hs22)
	checkAll(t, d2, []int{10, 30, 40})

	hs33 := InitWith(90)
	u3 := hs1.Union(hs33)
	checkAll(t, u3, []int{10, 20, 30, 40, 90})
	i3 := hs1.Intersection(hs33)
	checkAll(t, i3, []int{})
	d3 := hs1.Difference(hs33)
	checkAll(t, d3, []int{10, 20, 30, 40})

	hs44 := InitWith(20, 30, 50, 60)
	u4 := hs1.Union(hs44)
	checkAll(t, u4, []int{10, 20, 30, 40, 50, 60})
	i4 := hs1.Intersection(hs44)
	checkAll(t, i4, []int{20, 30})
	d4 := hs1.Difference(hs44)
	checkAll(t, d4, []int{10, 40})
}
