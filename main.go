package main

import "GinIM/server"

func main() {
	// server:=server.NewServer("192.168.10.153",8888)
	server := server.NewServer("127.0.0.1", 8888)
	server.Start()
}
