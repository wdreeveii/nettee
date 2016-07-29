package main

import (
	"fmt"
	"time"
)

func main() {
	in := make(chan []byte, 100)
	out := make(chan []byte)

	go func(src chan []byte) {
		for {
			data := <-src
			fmt.Print(string(data))
		}
	}(in)

	c, err := NewTeeClient("sadc_ts9:4039", 1*time.Minute, in, out)
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(10 * time.Minute)

	c.Close()
}
