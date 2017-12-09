package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/chinahdkj/rediqueue"
)

func main() {

	addr := ":6300"

	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	m, err := rediqueue.RunAddr(addr, 10)

	if err != nil {
		panic(err)
	}

	fmt.Println("rediqueue at " + addr)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c

	log.Println("rediqueue on exiting...")

	m.Save()
}
