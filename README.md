# MsgConn (Go)

[![GoDoc](https://godoc.org/github.com/boljen/msgconn?status.svg)](http://godoc.org/github.com/boljen/msgconn)
[![Build Status](https://travis-ci.org/boljen/msgconn.svg)](https://travis-ci.org/boljen/msgconn)

Package msgconn implements a wrapper that encodes and decodes variable-length
messages to and from a bytestream.

## Install

    go get github.com/boljen/msgconn
    go get gopkg.in/boljen/msgconn.v1

## Example

    // Creating a new MsgConn instance
    tcpConn, _ := ln.AcceptTCP()
    conn := &MsgConn{
            Reader: tcpConn,
            Writer: tcpConn,
            Closer: tcpConn,
    }

    // Reading and writing messages
    msg, _ := conn.ReadMessage()
    conn.WriteMessage([]byte("message"))

## License

This package is released under the MIT license.
