package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func usage() {
	fmt.Printf("usage: %s <host:port>\n", os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	listener, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		fmt.Printf("error: %s\n", err)
		usage()
	}
	defer listener.Close()
	log.Printf("listening on %s", os.Args[1])

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error accepting connection: %s\n", err)
		}
		go handleConn(conn)
	}
}

// handleConn run on its goroutine
func handleConn(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	log.Printf("%s> connection accepted", addr)

	defer conn.Close()
	defer log.Printf("%s> connection closed", addr)

	for {
		written, err := io.Copy(conn, conn)
		if err != nil {
			log.Printf("%s>error writing: %s", addr, err)
			break
		}
		if written == 0 {
			log.Printf("%s> done writing", addr)
			break
		}
	}
}
