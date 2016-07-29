package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	in := make(chan []byte, 10000)
	out := make(chan []byte)

	go func(src chan []byte) {
		for {
			data := <-src
			fmt.Print(string(data))
		}
	}(in)

	var s *TeeServer
	var c *TeeClient
	var err error

	s, err = NewTeeServer(":8080", in, out)
	if err != nil {
		fmt.Println(err)
	}

	c, err = NewTeeClient("sadc_ts9:4039", 1*time.Minute, in, out)
	if err != nil {
		fmt.Println(err)
	}

	e := make(chan os.Signal, 1)
	signal.Notify(e, syscall.SIGHUP, syscall.SIGINT)

SignalLoop:
	for {
		sig := <-e
		switch sig {
		case syscall.SIGHUP:
			// reload configs
			fmt.Println("sighup")
		case syscall.SIGINT:
			break SignalLoop
		}
	}
	c.Close()
	s.Close()

	close(out)
}
