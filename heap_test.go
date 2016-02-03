// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bheap

import (
	"container/heap"
	"math/rand"
	"os"
	"runtime/debug"
	"testing"
)

type myHeap []int

func (h *myHeap) Less(i, j int) bool {
	return (*h)[i] < (*h)[j]
}

func (h *myHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *myHeap) Len() int {
	return len(*h)
}

func (h *myHeap) Pop() (v interface{}) {
	*h, v = (*h)[:h.Len()-1], (*h)[h.Len()-1]
	return
}

func (h *myHeap) Push(v interface{}) {
	*h = append(*h, v.(int))
}

func (h myHeap) verify(t *testing.T, bh *BHeap) {
	for u := 1 + rootIndex; u < h.Len(); u++ {
		p := bh.parent(uint64(u))
		if h.Less(u-rootIndex, int(p)-rootIndex) {
			debug.PrintStack()
			bh.Dot(os.Stderr, "verify", &h, func(i uint64) interface{} { return h[int(i)] })
			t.Fatalf("h[%d] => %d < h[%d] => %d broken", p, h[p], u, h[u])
		}
	}
}

func TestInit0(t *testing.T) {
	bh := New(4)
	h := new(myHeap)
	for i := 20; i > 0; i-- {
		bh.Push(h, 0) // all elements are the same
	}
	h.verify(t, bh)

	for i := 1; h.Len() > 0; i++ {
		x := bh.Pop(h).(int)
		h.verify(t, bh)
		if x != 0 {
			t.Errorf("%d.th pop got %d; want %d", i, x, 0)
		}
	}
}

func TestInit1(t *testing.T) {
	bh := New(4)
	h := new(myHeap)
	for i := 20; i > 0; i-- {
		bh.Push(h, i) // all elements are different
	}
	h.verify(t, bh)

	for i := 1; h.Len() > 0; i++ {
		x := bh.Pop(h).(int)
		h.verify(t, bh)
		if x != i {
			t.Errorf("%d.th pop got %d; want %d", i, x, i)
		}
	}
}

func Test(t *testing.T) {
	bh := New(4)
	h := new(myHeap)
	h.verify(t, bh)

	for i := 20; i > 10; i-- {
		bh.Push(h, i)
	}

	h.verify(t, bh)

	for i := 10; i > 0; i-- {
		bh.Push(h, i)
		h.verify(t, bh)
	}

	for i := 1; h.Len() > 0; i++ {
		x := bh.Pop(h).(int)
		h.verify(t, bh)
		if i < 20 {
			bh.Push(h, 20+i)
		}
		if x != i {
			t.Fatalf("%d.th pop got %d; want %d", i, x, i)
		}
	}
}

func TestRemove0(t *testing.T) {
	bh := New(0)
	h := new(myHeap)
	for i := 0; i < 10; i++ {
		h.Push(i)
	}
	h.verify(t, bh)

	for h.Len() > 0 {
		i := h.Len() - 1
		x := bh.Remove(h, uint64(i)).(int)
		if x != i {
			t.Errorf("Remove(%d) got %d; want %d", i, x, i)
		}
		h.verify(t, bh)
	}
}

func TestRemove1(t *testing.T) {
	bh := New(0)

	h := new(myHeap)
	for i := 0; i < 10; i++ {
		h.Push(i)
	}
	h.verify(t, bh)

	for i := 0; h.Len() > 0; i++ {
		x := bh.Remove(h, 0).(int)
		if x != i {
			t.Errorf("Remove(0) got %d; want %d", x, i)
		}
		h.verify(t, bh)
	}
}

func TestRemove2(t *testing.T) {
	N := 10

	bh := New(0)
	h := new(myHeap)
	for i := 0; i < N; i++ {
		h.Push(i)
	}
	h.verify(t, bh)

	m := make(map[int]bool)
	for h.Len() > 0 {
		m[bh.Remove(h, uint64((h.Len()-1)/2)).(int)] = true
		h.verify(t, bh)
	}

	if len(m) != N {
		t.Errorf("len(m) = %d; want %d", len(m), N)
	}
	for i := 0; i < len(m); i++ {
		if !m[i] {
			t.Errorf("m[%d] doesn't exist", i)
		}
	}
}

func BenchmarkDup(b *testing.B) {
	bh := New(0)
	const n = 10000
	h := make(myHeap, n)
	for i := 0; i < b.N; i++ {
		for j := 0; j < n; j++ {
			bh.Push(&h, 0) // all elements are the same
		}
		for h.Len() > 0 {
			bh.Pop(&h)
		}
	}
}

func TestFix(t *testing.T) {
	bh := New(0)

	h := new(myHeap)
	h.verify(t, bh)

	for i := 200; i > 0; i -= 10 {
		bh.Push(h, i)
	}
	h.verify(t, bh)

	if (*h)[0] != 10 {
		t.Fatalf("Expected head to be 10, was %d", (*h)[0])
	}
	(*h)[0] = 210
	bh.Fix(h, 0)
	h.verify(t, bh)

	for i := 100; i > 0; i-- {
		elem := rand.Intn(h.Len())
		if i&1 == 0 {
			(*h)[elem] *= 2
		} else {
			(*h)[elem] /= 2
		}
		bh.Fix(h, uint64(elem))
		h.verify(t, bh)
	}
}

func BenchmarkPushHeap(b *testing.B) {
	h := new(myHeap)
	for i := 0; i < b.N; i++ {
		heap.Push(h, rand.Intn(i+1))
	}
}

func BenchmarkPushBHeap(b *testing.B) {
	bh := New(0)
	h := new(myHeap)
	for i := 0; i < b.N; i++ {
		bh.Push(h, rand.Intn(i+1))
	}
}
