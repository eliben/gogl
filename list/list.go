// Package list implements a doubly-linked list.
package list

import (
	"fmt"
	"iter"
)

// List is a doubly-linked list. Create new lists with [New] and then use
// the list's methods to interact with it.
type List[T any] struct {
	front  *Node[T]
	back   *Node[T]
	length int
}

// Node represents a node in the linked list; it holds a generic value, and
// can be used as a handle into the list for insertions/removals.
type Node[T any] struct {
	Value T

	next, prev *Node[T]
}

// New creates a new, empty linked-list.
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

// Len is the number of elements in the linked list. O(1)
func (lst *List[T]) Len() int {
	return lst.length
}

// Front returns the first node in the list.
func (lst *List[T]) Front() *Node[T] {
	if lst.length == 0 {
		return nil
	}
	return lst.front.next
}

// Back returns the last node in the list.
func (lst *List[T]) Back() *Node[T] {
	if lst.length == 0 {
		return nil
	}
	return lst.back.prev
}

// Next returns the next node in the list after `node`.
func (lst *List[T]) Next(node *Node[T]) *Node[T] {
	if nxt := node.next; nxt != lst.back {
		return nxt
	}
	return nil
}

// Prev returns the previous node in the list before `node`.
func (lst *List[T]) Prev(node *Node[T]) *Node[T] {
	if prv := node.prev; prv != lst.front {
		return prv
	}
	return nil
}

// InsertFront inserts a new node with the given value at the front of the list.
func (lst *List[T]) InsertFront(val T) {
	lst.InsertAfter(lst.front, val)
}

// InsertBack inserts a new node with the given value at the back of the list.
func (lst *List[T]) InsertBack(val T) {
	oldLast := lst.back.prev
	lst.InsertAfter(oldLast, val)
}

// InsertAfter inserts a new node with the given value after `node`.
func (lst *List[T]) InsertAfter(node *Node[T], val T) *Node[T] {
	newNode := &Node[T]{
		Value: val,
		next:  node.next,
		prev:  node,
	}
	newNode.next.prev = newNode
	newNode.prev.next = newNode
	lst.length++
	return newNode
}

// InsertBefore inserts a new node with the given value before `node`.
func (lst *List[T]) InsertBefore(node *Node[T], val T) *Node[T] {
	prev := node.prev
	return lst.InsertAfter(prev, val)
}

// Remove removes the given node from the list.
func (lst *List[T]) Remove(node *Node[T]) {
	node.prev.next = node.next
	node.next.prev = node.prev
	node.next = nil
	node.prev = nil
	lst.length--
}

// Values returns an iterator over all the values in the list.
func (lst *List[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		for node := lst.front.next; node != lst.back; node = node.next {
			if !yield(node.Value) {
				return
			}
		}
	}
}

// Nodes returns an iterator over all the nodes in the list.
func (lst *List[T]) Nodes() iter.Seq[*Node[T]] {
	return func(yield func(*Node[T]) bool) {
		for node := lst.front.next; node != lst.back; node = node.next {
			if !yield(node) {
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
