package msgconn

// Allocator allocates byteslices into memory.
type Allocator interface {
	Allocate(len, cap int) []byte
}

// MakeAllocator merely wraps the make([]byte, len, cap) statement.
// For most use cases this will be the only allocator needed.
var MakeAllocator Allocator = makeAllocator{}

// DefaultAllocator returns the default allocator instance
// which is an instance of MakeAllocator.
func DefaultAllocator() Allocator {
	return MakeAllocator
}

type makeAllocator struct{}

func (makeAllocator) Allocate(len int, cap int) []byte {
	return make([]byte, len, cap)
}
