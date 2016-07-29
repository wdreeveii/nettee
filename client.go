package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

type TeeClient struct {
	address string
	timeout time.Duration

	in  chan<- []byte
	out <-chan []byte

	conn net.Conn
	stop chan chan bool
}

func NewTeeClient(address string, timeout time.Duration, in chan<- []byte, out <-chan []byte) (*TeeClient, error) {
	t := new(TeeClient)
	t.stop = make(chan chan bool, 1)
	t.address = address
	t.timeout = timeout
	t.in = in
	t.out = out

	go t.handleConn()

	return t, nil
}

func (t *TeeClient) dial() *bufio.Reader {
	fmt.Println("dial..")
	var err error

	t.conn, err = net.DialTimeout("tcp", t.address, t.timeout)
	for err != nil {
		fmt.Println(err)

		time.Sleep(1 * time.Second)

		t.conn, err = net.DialTimeout("tcp", t.address, t.timeout)
	}

	r := bufio.NewReader(t.conn)

	return r
}

func (t *TeeClient) handleConn() {
	var err error
	r := t.dial()

	for {

		_, err = r.Peek(1)
		if err != nil {
			select {
			case done := <-t.stop:
				done <- true
				t.conn.Close()
				return
			default:
				r = t.dial()
			}
		}
		select {
		case output := <-t.out:
			t.conn.Write(output)
		default:
			var data []byte
			t.conn.SetDeadline(time.Now().Add(t.timeout))
			c, err := r.ReadByte()
			data = append(data, c)
			if err != nil {
				fmt.Println(err)
				t.conn.Close()
			}
			t.in <- data

		}
	}
}

func (t *TeeClient) Close() {
	done := make(chan bool)
	t.stop <- done
	t.conn.Close()
	<-done
	return
}
