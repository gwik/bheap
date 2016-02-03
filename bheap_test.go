package bheap

import "testing"

func TestNew(t *testing.T) {

	for _, s := range []int{505, 512} {

		bh := New(s)

		if bh.pageSize != 512 {
			t.Fatalf("expected page size of 512, was: %d", bh.pageSize)
		}

		if bh.pageMask != 0x1ff {
			t.Fatalf("expect mask to be 0xff was %d", bh.pageMask)
		}

		if bh.pageShift != 9 {
			t.Fatalf("expect shift to be 9 was %d", bh.pageShift)
		}

	}

}
