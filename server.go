package main

import (
	"errors"
	"fmt"
	"net"
)

type TeeServer struct {
	address         string
	in              <-chan []byte
	out             chan<- []byte
	listener        net.Listener
	stop            chan chan bool
	connectionIndex int
	connections     map[int]TeeServerClient
}

func NewTeeServer(address string, in <-chan []byte, out chan<- []byte) (*TeeServer, error) {
	t := new(TeeServer)

	t.address = address
	t.in = in
	t.out = out

	t.stop = make(chan chan bool, 1)
	t.connections = make(map[int]TeeServerClient)

	ln, err := net.Listen("tcp", t.address)
	if err != nil {
		return nil, err
	}

	t.listener = ln

	go t.pubsub()
	go t.acceptConns()

	return t, nil
}

func (t *TeeServer) newConnectionId() (int, error) {
	idx := t.connectionIndex
	t.connectionIndex += 1

	var iterations int
	for _, exists := t.connections[idx]; exists; _, exists = t.connections[idx] {
		idx = t.connectionIndex
		t.connectionIndex += 1
		iterations += 1

		if iterations == 0 {
			return 0, errors.New("Maximum number of active connections reached.")
		}
	}

	return idx, nil
}

func (t *TeeServer) pubsub() {
	for c := range t.cmd {
		switch c.op {
		case pub:
			for _, v := range t.connections {
				v.toClient <- c.slice
			}
		case sub:
					idx, err := t.newConnectionId()
		if err != nil {
			conn.Close()
			continue
		}

		var c TeeServerClient = NewTeeServerClient(conn, idx)

		go c.handle()

		t.connections[idx] = c
			t.connections[c.connId] = cmd.teeServerClient
		case unsub:
			delete t.connections[c.connId]
		}
	}

	for _, v := range t.connections {
		v.Close()
	}
}

func (t *TeeServer) acceptConns() {

	for {
		conn, err := t.listener.Accept()
		if err != nil {

			select {
			case done := <-t.stop:
				close(t.cmd)
				done <- true
				return
			default:
				fmt.Println(err)

				continue
			}
		}

		t.cmd<- cmd{op:sub, conn: conn}

	}
}

type TeeServerClient struct {
	toClient   chan []byte
	fromClient chan []byte

	stop chan chan bool

	conn   net.Conn
	connId int
}

func NewTeeServerClient(conn net.Conn, connId int) TeeServerClient {
	var c TeeServerClient
	c.toClient = make(chan []byte)
	c.fromClient = make(chan []byte)

	c.stop = make(chan chan bool)

	c.conn = conn
	c.connId = connId
	return c
}

func (c *TeeServerClient) output() {
	for slice := range c.toClient {
		c.conn.Write(slice)
	}
}

func (c *TeeServerClient) handle() {

	for {
		select {
		case done := <-c.stop:
			done <- true
			return
		default:

		}
	}
	c.conn.Write([]byte("Hello World."))
}

func (c *TeeServerClient) Close() {
	c.conn.Close()
}

func (t *TeeServer) Close() {
	done := make(chan bool)
	t.stop <- done
	t.listener.Close()
	<-done
	return
}
