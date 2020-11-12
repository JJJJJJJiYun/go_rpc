package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/JJJJJJJiYun/go_rpc/codec"
	"github.com/JJJJJJJiYun/go_rpc/server"
)

func main() {
	addr := make(chan string)
	go startServer(addr)
	conn, _ := net.Dial("tcp", <-addr)
	defer func() { _ = conn.Close() }()
	_ = json.NewEncoder(conn).Encode(server.DefaultOption)
	cc := codec.NewCodecFuncMap[codec.JsonType](conn)
	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Test",
			Seq:           uint64(i),
		}
		_ = cc.Write(h, fmt.Sprintf("go_rpc req %d", h.Seq))
		_ = cc.ReadHeader(h)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("reply: ", reply)
	}
}

func startServer(addr chan string) {
	l, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal(err)
	}
	addr <- l.Addr().String()
	server.Accept(l)
}
