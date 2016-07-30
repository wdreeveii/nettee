package main

import (
	"flag"
	"fmt"
	"github.com/wdreeveii/nettee/nettee"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var localAddress string
var remoteAddress string
var timeout time.Duration

func init() {
	flag.StringVar(&localAddress, "localAddress", "", "local address")
	flag.StringVar(&remoteAddress, "remoteAddress", "", "remote address")
	flag.DurationVar(&timeout, "timeout", time.Hour, "remote connection timeout")
}

func main() {
	flag.Parse()

	if localAddress == "" || remoteAddress == "" {
		flag.PrintDefaults()
		return
	}

	t, err := nettee.NewTee(localAddress, remoteAddress, timeout)
	if err != nil {
		panic(err)
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

	t.Close()
}
