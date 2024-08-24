// Package list implements a doubly-linked list.
package list

import (
	"fmt"
	"iter"
)

type Node[T any] struct {
	next, prev *Node[T]
	Value      T
}

type List[T any] struct {
	front  *Node[T]
	back   *Node[T]
	length int
}

func New[T any]() *List[T] {
	// A List always has two allocated sentinel nodes: front points to the
	// first, and back points at the last. The length of the list remains
	// 0 as long as there are only sentinel nodes in it.
	lst := &List[T]{
		front:  &Node[T]{},
		back:   &Node[T]{},
		length: 0,
	}
	lst.front.next = lst.back
	lst.back.prev = lst.front
	return lst
}

func (lst *List[T]) Len() int {
	return lst.length
}

func (lst *List[T]) Front() *Node[T] {
	if lst.length == 0 {
		return nil
	}
	return lst.front.next
}

func (lst *List[T]) Back() *Node[T] {
	if lst.length == 0 {
		return nil
	}
	return lst.back.prev
}

// InsertFront inserts a new node with the given value at the front of the list.
func (lst *List[T]) InsertFront(val T) {
	oldFirst := lst.front.next
	lst.front.next = &Node[T]{
		next:  oldFirst,
		prev:  lst.front,
		Value: val,
	}
	oldFirst.prev = lst.front.next
	lst.length += 1
}

// InsertBack inserts a new node with the given value at the back of the list.
func (lst *List[T]) InsertBack(val T) {
	oldLast := lst.back.prev
	lst.back.prev = &Node[T]{
		next:  lst.back,
		prev:  oldLast,
		Value: val,
	}
	oldLast.next = lst.back.prev
	lst.length += 1
}

// All returns an iterator over all the values in the list.
func (lst *List[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for node := lst.front.next; node != lst.back; node = node.next {
			if !yield(node.Value) {
				return
			}
		}
	}
}

func (lst *List[T]) debugPrint() {
	fmt.Println("-----------------------")
	for n := lst.front; n != nil; n = n.next {
		var specialName string
		if n == lst.front {
			specialName = " [FRONT]"
		} else if n == lst.back {
			specialName = " [BACK]"
		}
		fmt.Printf("| Node addr=%p%v:\n", n, specialName)
		fmt.Printf("|   value=%v    next=%p    prev=%p\n", n.Value, n.next, n.prev)
	}
	fmt.Println("-----------------------")
}
