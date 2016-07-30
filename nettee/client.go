package nettee

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

func NewTeeClient(address string, timeout time.Duration, in chan<- []byte, out <-chan []byte) *TeeClient {
	t := new(TeeClient)
	t.stop = make(chan chan bool, 1)
	t.address = address
	t.timeout = timeout
	t.in = in
	t.out = out

	go t.handleConn()

	return t
}

func (t *TeeClient) dial(r *bufio.Reader) (*bufio.Reader, error) {
	fmt.Println("dial..")
	var err error

	t.conn, err = net.DialTimeout("tcp", t.address, t.timeout)
	if err != nil {
		return r, err
	}

	r = bufio.NewReader(t.conn)

	return r, nil
}

func (t *TeeClient) handleWrite() {
	for m := range t.out {
		if t.conn != nil {
			t.conn.Write(m)
		}
	}
}

func (t *TeeClient) handleConn() {
	go t.handleWrite()

	var err error
	var r *bufio.Reader = new(bufio.Reader)

	for {

		_, err = r.Peek(1)
		if err != nil {
			select {
			case done := <-t.stop:
				done <- true
				return
			default:
				r, err = t.dial(r)
				if err != nil {
					fmt.Println(err)
					time.Sleep(time.Second)
					continue
				}
			}
		}

		var data []byte
		t.conn.SetDeadline(time.Now().Add(t.timeout))
		c, err := r.ReadByte()
		data = append(data, c)
		if err != nil {
			fmt.Println(err)
			t.conn.Close()
			continue
		}
		t.in <- data

	}
}

func (t *TeeClient) Close() {
	done := make(chan bool)
	t.stop <- done
	if t.conn != nil {
		t.conn.Close()
	}
	<-done
	return
}
