package msgconn

import (
	"errors"
	"io"
)

// Errors that can be returned by the various serialization methods.
var (
	ErrPrefixBufferSize = errors.New("MsgConn: Prefix Buffer has an invalid size, must be exactly 3 bytes")
	ErrMsgSize          = errors.New("MsgConn: Message has illegal size")
)

// MaxLegalSize is the maximum legal size of a message and is equal to the
// highest unsigned integer that can be stored inside 3 bytes.
const MaxLegalSize = 256 * 256 * 256

// ParsePrefix parses the length prefix from a byteslice.
func ParsePrefix(b []byte) int {
	return int(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16)
}

// WritePrefix writes the length prefix to a byteslice.
func WritePrefix(b []byte, l int) {
	v := uint32(l) // Does this matter?
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
}

// ReadAll reads from io.Reader until byteslice b is entirely filled.
// An error will be returned if it failed reading all of the data.
func ReadAll(r io.Reader, b []byte) error {
	l := len(b)
	index := 0
	for index < l {
		done, err := r.Read(b[index:])
		if err != nil {
			return err
		}
		index += done
	}
	return nil
}

// ReadMessage reads a message from r into a byteslice allocated by a.
// It will always return a slice if one has been allocated, even if
// an error has occured. In the latter case the slice might contain
// partial data.
// c is a byteslice to which the length prefix will be read and must
// be 3 bytes in size.
func ReadMessage(r io.Reader, a Allocator, c []byte, max int) ([]byte, error) {
	if len(c) != 3 {
		return nil, ErrPrefixBufferSize
	}
	if err := ReadAll(r, c); err != nil {
		return nil, err
	}
	l := ParsePrefix(c)
	if l > max || l > MaxLegalSize || l <= 0 {
		return nil, ErrMsgSize
	}
	data := a.Allocate(int(l), int(l))
	return data, ReadAll(r, data)
}

// WriteMessage writes a length prefix and message m to writer w.
// It caches the length prefix to byteslice c before writing it out.
func WriteMessage(w io.Writer, m, c []byte, max int) error {
	if len(c) != 3 {
		return ErrPrefixBufferSize
	}
	if len(m) > max || len(m) > MaxLegalSize || len(m) <= 0 {
		return ErrMsgSize
	}
	WritePrefix(c, len(m))
	if _, err := w.Write(c); err != nil {
		return err
	}
	_, err := w.Write(m)
	return err
}
