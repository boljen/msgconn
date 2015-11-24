package msgconn

import (
	"bytes"
	"errors"
	"testing"
)

type flusher struct{ E error }

func (f *flusher) Flush() error { return f.E }

type closer struct{ E error }

func (c *closer) Close() error { return c.E }

func TestMsgConnMaxSize(t *testing.T) {
	conn := &MsgConn{}
	if conn.maxInSize() != DefaultMaxMsgSize || conn.maxOutSize() != DefaultMaxMsgSize {
		t.Fatal("wrong default sizes")
	}

	conn.MaxInSize = 10
	conn.MaxOutSize = 20
	if conn.maxInSize() != 10 || conn.maxOutSize() != 20 {
		t.Fatal("wrong custom sizes")
	}
}
func TestMsgConnGetAllocator(t *testing.T) {
	conn := &MsgConn{}
	if conn.GetAllocator() != MakeAllocator {
		t.Fatal("didn't return the default allocator")
	}
	a := &makeAllocator{}
	conn.Allocator = a
	if conn.GetAllocator() == MakeAllocator || conn.GetAllocator() != a {
		t.Fatal("returned the wrong allocator")
	}
}

func TestMsgConnReadMessage(t *testing.T) {
	data := []byte{4, 0, 0, 116, 101, 115, 116}
	reader := bytes.NewReader(data)
	conn := &MsgConn{
		Reader:     reader,
		MaxInSize:  5,
		MaxOutSize: 5,
	}

	if msg, err := conn.ReadMessage(); err != nil {
		t.Fatal("unexpected error", err)
	} else if string(msg) != "test" {
		t.Fatal("expected 'test'")
	}
}

func TestMsgConnWriteMessage(t *testing.T) {
	errFlush := errors.New("flush error")

	w := new(bytes.Buffer)
	f := &flusher{E: errFlush}
	conn := &MsgConn{
		Writer:     w,
		Flusher:    f,
		MaxInSize:  5,
		MaxOutSize: 5,
	}

	// Test a successfull write
	data := []byte{116, 101, 115, 116}
	if err := conn.WriteMessage(data, false); err != nil {
		t.Fatal("expected nil, got ", err)
	} else if bytes.Compare(w.Bytes(), []byte{4, 0, 0, 116, 101, 115, 116}) != 0 {
		t.Fatal("expected bytes to compare")
	}

	// Test a successfull write with flush, validate using mock flush error.
	if err := conn.WriteMessage(data, true); err != errFlush {
		t.Fatal("expected errFlush, got ", err)
	}

	// Test an unsuccessfull write
	conn.MaxOutSize = 1
	if err := conn.WriteMessage(data, true); err != ErrMsgSize {
		t.Fatal("expected ErrMsgSize, got ", err)
	}
}

func TestMsgConnFlush(t *testing.T) {
	errFlush := errors.New("flush error")

	conn := &MsgConn{}

	if err := conn.Flush(); err != nil {
		t.Fatal("unexpected error", err)
	}

	conn.Flusher = &flusher{E: errFlush}
	if err := conn.Flush(); err != errFlush {
		t.Fatal("expected errFlush")
	}
}

func TestMsgConnClose(t *testing.T) {
	errFlush := errors.New("flush error")
	errClose := errors.New("close error")

	cl := &closer{}
	fl := &flusher{}
	conn := &MsgConn{}

	if err := conn.Close(); err != nil {
		t.Fatal("unexpected error", err)
	}
	conn.Closer = cl
	conn.Flusher = fl
	cl.E = errClose
	if err := conn.Close(); err != errClose {
		t.Fatal("unexpected error", err)
	}

	fl.E = errFlush
	if err := conn.Close(); err != errFlush {
		t.Fatal("unexpected error", err)
	}
}
