package msgconn

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"testing/iotest"
)

type MockWriter struct {
	I int
	E error
	U func(*MockWriter)
}

func (m *MockWriter) Write([]byte) (int, error) {
	i, e := m.I, m.E
	if m.U != nil {
		m.U(m)
	}
	return i, e
}

func TestReadWritePrefix(t *testing.T) {
	data := make([]byte, 3)
	data[0] = 100
	data[1] = 1
	if ParsePrefix(data) != 356 {
		t.Fatal("wrong prefix")
	}

	data2 := make([]byte, 3)
	data2[2] = 255

	WritePrefix(data2, 356)
	if data2[0] != 100 || data[1] != 1 || data[2] != 0 {
		t.Fatal("wrong prefix")
	}
}

func TestReadAll(t *testing.T) {
	orw := bytes.NewReader([]byte("tester"))
	rw := iotest.OneByteReader(orw)
	data := []byte("aaaaaa")

	if err := ReadAll(rw, data); err != nil {
		t.Fatal("unexpected error", err)
	}

	rw = iotest.DataErrReader(orw)

	if err := ReadAll(rw, data); err != io.EOF {
		t.Fatal("expected io.EOF")
	}

}

func TestReadMessage(t *testing.T) {
	var r io.Reader
	a := DefaultAllocator()
	data := []byte{4, 0, 0, 116, 101, 115, 116}
	if ParsePrefix(data) != 4 {
		t.Fatal("wrong prefix")
	}

	max := 10
	r = bytes.NewReader(data)

	/* Test Prefix Buffer Size */
	p := []byte{0, 0}
	if data, err := ReadMessage(r, a, p, max); err != ErrPrefixBufferSize || data != nil {
		t.Fatal("expected ErrPrefixBufferSize")
	}
	p = []byte{0, 0, 0, 0}
	if data, err := ReadMessage(r, a, p, max); err != ErrPrefixBufferSize || data != nil {
		t.Fatal("expected ErrPrefixBufferSize")
	}
	p = []byte{0, 0, 0}

	/* Test Data Errors */
	r = iotest.DataErrReader(bytes.NewReader([]byte{10, 0}))
	if data, err := ReadMessage(r, a, p, max); err != io.EOF || data != nil {
		t.Fatal("expected io.EOF", err, data)
	}

	r = bytes.NewReader(data)
	if data, err := ReadMessage(bytes.NewReader(data), a, p, 3); err != ErrMsgSize || data != nil {
		t.Fatal("expected ErrMsgSize")
	}

	if data, err := ReadMessage(bytes.NewReader(data), a, p, 4); err != nil || string(data) != "test" {
		t.Fatal("expected ErrMsgSize")
	}
}

var errTestWriter = errors.New("test writer error")

func TestWriteMessage(t *testing.T) {
	var w io.Writer
	m := []byte("tes")
	max := 10

	/* Test prefix */
	c := []byte{0, 0}
	if err := WriteMessage(w, m, c, max); err != ErrPrefixBufferSize {
		t.Fatal("expected ErrPrefixBufferSize")
	}
	c = []byte{0, 0, 0, 0}
	if err := WriteMessage(w, m, c, max); err != ErrPrefixBufferSize {
		t.Fatal("expected ErrPrefixBufferSize")
	}
	c = []byte{0, 0, 0}

	/* Test size */
	if err := WriteMessage(w, m, c, 2); err != ErrMsgSize {
		t.Fatal("expected ErrMsgSize")
	}

	/* Test write errors */
	mw := &MockWriter{
		I: 0,
		E: errTestWriter,
	}
	if err := WriteMessage(mw, m, c, max); err != errTestWriter {
		t.Fatal("expected errTestWriter")
	}

	/* Test write error fallthrough */
	mw.E = nil
	mw.I = 3
	mw.U = func(m *MockWriter) {
		mw.E = errTestWriter
	}
	if err := WriteMessage(mw, m, c, max); err != errTestWriter {
		t.Fatal("expected errTestWriter")
	}
}
