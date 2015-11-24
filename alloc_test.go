package msgconn

import "testing"

func TestDefaultAllocator(t *testing.T) {
	alloc := DefaultAllocator()
	data := alloc.Allocate(10, 20)
	if len(data) != 10 || cap(data) != 20 {
		t.Fatal("wrong allocation")
	}
}
