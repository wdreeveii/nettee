package nettee

import (
	"time"
)

type Tee struct {
	localAddress  string
	remoteAddress string
	remoteTimeout time.Duration

	in  chan []byte
	out chan []byte

	client *TeeClient
	server *TeeServer
}

func NewTee(local string, remote string, timeout time.Duration) (*Tee, error) {
	t := new(Tee)

	t.localAddress = local
	t.remoteAddress = remote
	t.remoteTimeout = timeout

	t.in = make(chan []byte, 10000)
	t.out = make(chan []byte)

	var err error

	t.server, err = NewTeeServer(t.localAddress, t.in, t.out)
	if err != nil {
		return nil, err
	}

	t.client = NewTeeClient(t.remoteAddress, t.remoteTimeout, t.in, t.out)

	return t, nil

}

func (t *Tee) Close() {
	t.server.Close()
	t.client.Close()
	close(t.out)
}
