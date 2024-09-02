package btree

import (
	"testing"
)

func intCmp(a, b int) int {
	return a - b
}

func TestBasic(t *testing.T) {
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

	bt.renderDotToImage("bt.png")
}
