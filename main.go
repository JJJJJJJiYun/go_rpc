package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/JJJJJJJiYun/go_rpc/client"
	"github.com/JJJJJJJiYun/go_rpc/server"
)

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)
	c, err := client.Dial("tcp", <-addr)
	defer func() { _ = c.Close() }()
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := fmt.Sprintf("gorpc req %d", i)
			var reply string
			if err := c.Call("Foo.Test", args, &reply); err != nil {
				log.Fatal("call Foo.Test err: ", err)
			}
			log.Println("reply: ", reply)
		}(i)
	}
	wg.Wait()
}

func startServer(addr chan string) {
	l, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal(err)
	}
	addr <- l.Addr().String()
	server.Accept(l)
}
