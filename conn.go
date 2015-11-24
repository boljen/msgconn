// Package msgconn implements a wrapper that encodes and decodes variable-length
// messages to and from a bytestream, typically a connection.
//
// MsgConn Creation
//
// Creating a MsgConn is as easy as creating any other struct and
// setting some of it's fields. Only the Reader and Writer are
// required and only when using the ReadMessage and WriteMessage
// methods respectively.
//
//     conn := &MsgConn{
//         Reader: r,
//         Writer: w,
//     }
//
// Bear in mind that the Reader, Writer and Closer fields can all
// be fulfilled by the same net.Conn instance.
//
//     tcpConn, _ := ln.AcceptTCP()
//     conn := &MsgConn{
//         Reader: tcpConn,
//         Writer: tcpConn,
//         Closer: tcpConn,
//     }
//
// MsgConn IO
//
// Sending and receiving messages between MsgConn instances is easy.
//
//     // Create a new duplex MsgConn link.
//     c1, c2 := &MsgConn{...}, &MsgConn{...}
//
//     // Write the message and flush.
//     err := c1.WriteMessage([]byte("test"), true)
//
//     // Read the message.
//     msg, err := c2.ReadMessage()
//
//
// Encoding
//
// The messages are framed by adding a fixed-size length prefix.
// The size of the prefix has been set to 3 bytes allowing for
// message sizes up to 2^24 bytes. The actual length is encoded
// as an unsigned integer using LittleEndian byte order.
//
//    // Example:
//    msg := []byte("test")
//    [ 4, 0, 0, 116, 101, 115, 116]
package msgconn

import "io"

// DefaultMaxMsgSize is the default maximum message size that
// can be sent or received by a MsgConn instance.
const DefaultMaxMsgSize = 256

// Flusher wraps around the Flush() method.
type Flusher interface {
	Flush() error
}

// MsgConn represents a message-based connection.
type MsgConn struct {
	Allocator Allocator

	// These are the maximum sizes of incoming and outbound messages.
	// The DefaultMaxMsgSize will be used when they are not set.
	// MaxLegalSize will be enforced separately.
	MaxInSize  int
	MaxOutSize int

	// These are the various interfaces that are used to process messages.
	// Reader and Writer are required, Closer and Flusher are optional.
	Reader  io.Reader
	Writer  io.Writer
	Closer  io.Closer
	Flusher Flusher

	// TODO: cheaper if [6]byte?
	rp [3]byte // read prefix
	wp [3]byte // write prefix
}

func (mc *MsgConn) maxInSize() int {
	if mc.MaxInSize <= 0 {
		return DefaultMaxMsgSize
	}
	return mc.MaxInSize
}

func (mc *MsgConn) maxOutSize() int {
	if mc.MaxOutSize <= 0 {
		return DefaultMaxMsgSize
	}
	return mc.MaxOutSize
}

// GetAllocator returns the current memory allocator of the connection.
// It will return the default allocator instance if no allocator is set.
func (mc *MsgConn) GetAllocator() Allocator {
	if mc.Allocator == nil {
		return MakeAllocator
	}
	return mc.Allocator
}

// ReadMessage reads a message from the underlying Reader.
// If a byteslice is allocated through Allocator then it will
// be returned even if an error occurs.
func (mc *MsgConn) ReadMessage() ([]byte, error) {
	return ReadMessage(mc.Reader, mc.GetAllocator(), mc.rp[:], mc.maxInSize())
}

// WriteMessage writes the message to the underlying Writer.
// If flush is true, the underlying buffer will be flushed but
// only if a Flusher has been set.
func (mc *MsgConn) WriteMessage(msg []byte, flush bool) error {
	if err := WriteMessage(mc.Writer, msg, mc.wp[:], mc.maxOutSize()); err != nil {
		return err
	}
	if !flush || mc.Flusher == nil {
		return nil
	}
	return mc.Flusher.Flush()
}

// Flush flushing any remaining pending data.
func (mc *MsgConn) Flush() error {
	if mc.Flusher == nil {
		return nil
	}
	return mc.Flusher.Flush()
}

// Close first flushes the connection and then calls closeµ
// on the underlying closer instance, if any.
func (mc *MsgConn) Close() error {
	err := mc.Flush()
	if mc.Closer == nil {
		return err
	}
	if err != nil {
		mc.Closer.Close()
		return err
	}
	return mc.Closer.Close()
}
