package msgconn

import "net"

func Example() {
	// Setup the duplex MsgConn connection and define messages.
	msg1, msg2 := "m1_to_m2", "m2_to_m1"
	c1, c2 := net.Pipe()
	m1 := &MsgConn{Reader: c1, Closer: c1, Writer: c1}
	m2 := &MsgConn{Reader: c2, Closer: c2, Writer: c2}

	// Read msg1 and send msg2 through m2.
	go func() {
		if msg, err := m2.ReadMessage(); err != nil || string(msg) != msg1 {
			panic("failed reading msg1")
		}
		if err := m2.WriteMessage([]byte(msg2), true); err != nil {
			panic("failed writing msg2")
		}
	}()

	// Write msg1 and read msg2 through m1
	if err := m1.WriteMessage([]byte(msg1), true); err != nil {
		panic("failed writing msg1")
	}
	if msg, err := m1.ReadMessage(); err != nil || string(msg) != msg2 {
		panic("failed reading msg2")
	}
}
