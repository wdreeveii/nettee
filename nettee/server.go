package nettee

import (
	"bufio"
	"errors"
	"fmt"
	"net"
)

type connectionIndex int

type TeeServer struct {
	address string

	in  <-chan []byte
	out chan<- []byte

	listener net.Listener

	stop    chan chan bool
	cleanup chan chan bool
	sub     chan net.Conn
	unsub   chan connectionIndex

	conIdx      connectionIndex
	connections map[connectionIndex]TeeServerClient
}

func NewTeeServer(address string, in <-chan []byte, out chan<- []byte) (*TeeServer, error) {
	t := new(TeeServer)

	t.address = address
	t.in = in
	t.out = out

	t.stop = make(chan chan bool, 1)
	t.cleanup = make(chan chan bool)
	t.sub = make(chan net.Conn)
	t.unsub = make(chan connectionIndex)
	t.connections = make(map[connectionIndex]TeeServerClient)

	ln, err := net.Listen("tcp", t.address)
	if err != nil {
		return nil, err
	}

	t.listener = ln

	go t.pubsub()
	go t.acceptConns()

	return t, nil
}

func (t *TeeServer) newConnectionId() (connectionIndex, error) {
	idx := t.conIdx
	t.conIdx += 1

	var iterations int
	for _, exists := t.connections[idx]; exists; _, exists = t.connections[idx] {
		idx = t.conIdx
		t.conIdx += 1
		iterations += 1

		if iterations == 0 {
			return 0, errors.New("Maximum number of active connections reached.")
		}
	}

	return idx, nil
}

func (t *TeeServer) pubsub() {
	var done chan bool

	for {
		if done != nil && len(t.connections) == 0 {
			done <- true
			return
		}

		select {
		case done = <-t.cleanup:
			for _, v := range t.connections {
				v.Close()
			}
		case data := <-t.in:
			for _, v := range t.connections {
				v.toClient <- data
			}
		case conn := <-t.sub:
			idx, err := t.newConnectionId()
			if err != nil {
				conn.Close()
				continue
			}

			var c TeeServerClient = NewTeeServerClient(conn, idx, t.out, t.unsub)

			t.connections[idx] = c
		case idx := <-t.unsub:
			delete(t.connections, idx)
		}

	}
}

func (t *TeeServer) acceptConns() {

	for {
		conn, err := t.listener.Accept()
		if err != nil {
			done := <-t.stop
			subdone := make(chan bool)
			t.cleanup <- subdone
			<-subdone
			done <- true
			return

		}

		t.sub <- conn

	}
}

func (t *TeeServer) Close() {
	done := make(chan bool)
	t.stop <- done
	t.listener.Close()
	<-done
	return
}

type TeeServerClient struct {
	toClient   chan []byte
	fromClient chan<- []byte

	stop  chan chan bool
	unsub chan<- connectionIndex

	conn   net.Conn
	connId connectionIndex
}

func NewTeeServerClient(conn net.Conn,
	connId connectionIndex,
	out chan<- []byte,
	unsub chan<- connectionIndex) TeeServerClient {

	var c TeeServerClient
	c.toClient = make(chan []byte)
	c.fromClient = out

	c.stop = make(chan chan bool)
	c.unsub = unsub
	c.conn = conn
	c.connId = connId

	go c.output()
	go c.handle()

	return c
}

func (c *TeeServerClient) output() {
	for slice := range c.toClient {
		c.conn.Write(slice)
	}
}

func (c *TeeServerClient) handle() {

	r := bufio.NewReader(c.conn)

	for {
		var data []byte
		d, err := r.ReadByte()
		data = append(data, d)
		if err != nil {
			fmt.Println(err)
			break
		}
		c.fromClient <- data
	}
	close(c.toClient)
	c.conn.Close()

	c.unsub <- c.connId
}

func (c *TeeServerClient) Close() {
	c.conn.Close()
	return
}
