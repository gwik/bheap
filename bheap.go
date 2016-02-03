package bheap

import (
	"container/heap"
	"fmt"
	"io"
)

type BHeap struct {
	pageSize  uint64
	pageMask  uint64
	pageShift uint64
}

type Interface heap.Interface

const (
	DefaultPageSize = 512
	rootIndex       = 1

	maxUint64 uint64 = 1<<64 - 1 // from math package but not imported
)

func New(pageSize int) *BHeap {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}

	var ps, s uint64
	for s = 1; (1 << s) < pageSize; s++ {
	}

	ps = 1 << s

	return &BHeap{
		pageSize:  ps,
		pageMask:  ps - 1,
		pageShift: s,
	}
}

func (bh *BHeap) parent(u uint64) uint64 {
	var v uint64

	po := u & bh.pageMask

	if u < bh.pageSize || po > 3 {
		v = (u & ^bh.pageMask) | (po >> 1)
	} else if po < 2 {
		v = (u - bh.pageSize) >> bh.pageShift
		v += v & ^(bh.pageMask >> 1)
		v |= bh.pageSize / 2
	} else {
		v = u - 2
	}
	return v
}

func (bh *BHeap) child(p uint64) (left uint64, right uint64) {

	if p > bh.pageMask && (p&(bh.pageMask-1)) == 0 {
		/* First two elements are magical except on the first page */
		left = p + 2
		right = p + 2
	} else if p&(bh.pageSize>>1) != 0 {
		/* The bottom row is even more magical */
		left = (p & ^bh.pageMask) >> 1
		left |= p & (bh.pageMask >> 1)
		left += 1
		uu := left << bh.pageShift
		left = uu
		if left == uu {
			right = left + 1
		} else {
			/*
			 * An unsigned is not big enough: clamp instead
			 * of truncating.  We do not support adding
			 * more than maxUint64 elements anyway, so this
			 * is without consequence.
			 */
			left = maxUint64
			right = maxUint64
		}
	} else {
		/* The rest is as usual, only inside the page */
		left = p + (p & bh.pageMask)
		right = left + 1
	}

	return left, right
}

func (bh *BHeap) up(h Interface, u uint64) uint64 {
	for u > rootIndex {
		p := bh.parent(u)
		if !h.Less(int(u)-rootIndex, int(p)-rootIndex) {
			break
		}
		h.Swap(int(u)-rootIndex, int(p)-rootIndex)
		u = p
	}
	return u
}

func (bh *BHeap) down(h Interface, u uint64) uint64 {

	for {
		v1, v2 := bh.child(u)

		if v1 >= uint64(h.Len())+rootIndex {
			return u
		}

		if v1 != v2 && v2 < uint64(h.Len())+rootIndex {
			if h.Less(int(v2)-rootIndex, int(v1)-rootIndex) {
				v1 = v2
			}
		}
		if h.Less(int(u)-rootIndex, int(v1)-rootIndex) {
			return u
		}
		h.Swap(int(u)-rootIndex, int(v1)-rootIndex)
		u = v1
	}

	return u
}

func (bh *BHeap) Push(h Interface, v interface{}) uint64 {
	h.Push(v)
	return bh.up(h, uint64(h.Len())-1+rootIndex)
}

func (bh *BHeap) Pop(h Interface) interface{} {
	h.Swap(0, h.Len()-1)
	v := h.Pop()
	bh.down(h, rootIndex)
	return v
}

func (bh *BHeap) Remove(h Interface, i uint64) interface{} {
	n := uint64(h.Len()) - 1
	if n != i {
		h.Swap(int(i), int(n))
		v := h.Pop()
		bh.down(h, i+rootIndex)
		bh.up(h, i+rootIndex)
		return v
	}
	return h.Pop()
}

func (bh *BHeap) Fix(h Interface, i uint64) {
	bh.down(h, bh.up(h, i+rootIndex))
}

func (bh *BHeap) dotChild(p uint64, w io.Writer, h Interface, valFn func(i uint64) interface{}) error {

	if _, err := w.Write([]byte(fmt.Sprintf("%d [label=\"%d|%d\"];\n", p, valFn(p-rootIndex), p))); err != nil {
		return err
	}

	v1, v2 := bh.child(p)
	if v1 >= uint64(h.Len())+rootIndex {
		return nil
	}

	if v1 != v2 && v2 < uint64(h.Len())+rootIndex {
		if h.Less(int(v2)-rootIndex, int(v1)-rootIndex) {
			v1, v2 = v2, v1
		}

		if _, err := w.Write([]byte(fmt.Sprintf("%d -> %d [color=black];\n", p, v1))); err != nil {
			return err
		}
		if _, err := w.Write([]byte(fmt.Sprintf("%d -> %d [color=red];\n", p, v2))); err != nil {
			return err
		}
		if err := bh.dotChild(v1, w, h, valFn); err != nil {
			return err
		}
		if err := bh.dotChild(v2, w, h, valFn); err != nil {
			return err
		}
	} else {
		color := "black"
		if v1 == v2 {
			color = "green"
		}
		if _, err := w.Write([]byte(fmt.Sprintf("%d -> %d [color=%s];\n", p, v1, color))); err != nil {
			return err
		}
		if err := bh.dotChild(v1, w, h, valFn); err != nil {
			return err
		}
	}
	return nil
}

func (bh *BHeap) Dot(w io.Writer, name string, h Interface, valFn func(i uint64) interface{}) error {

	if _, err := w.Write([]byte(fmt.Sprintf("\ndigraph %s {\nnode [shape=record];\n", name))); err != nil {
		return err
	}

	if err := bh.dotChild(rootIndex, w, h, valFn); err != nil {
		return err
	}

	if _, err := w.Write([]byte("\n}\n")); err != nil {
		return err
	}

	return nil

}
