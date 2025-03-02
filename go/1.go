package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

type ResponseType int

const (
	Malformed ResponseType = -1
	NotPrime               = 0
	Prime                  = 1
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

	buf := bufio.NewScanner(conn)
	for buf.Scan() {
		req := buf.Bytes()
		log.Printf("%s> --> %s", addr, req)

		rt := respType(req)
		var resp []byte
		switch rt {
		case NotPrime:
			resp = []byte(`{"method":"isPrime","prime":false}`)
			log.Printf("%s> <-- %s", addr, resp)
		case Prime:
			resp = []byte(`{"method":"isPrime","prime":true}`)
			log.Printf("%s> <-- %s", addr, resp)
		default:
			log.Printf("%s> <-- malformed", addr)
		}
		resp = append(resp, byte('\n'))

		if _, err := conn.Write(resp); err != nil {
			log.Fatalf("%s> error writing to conn: %s", addr, err)
		}
		if rt == Malformed {
			break
		}
	}
}

func respType(req []byte) ResponseType {
	m := map[string]any{}
	if err := json.Unmarshal(req, &m); err != nil {
		return Malformed
	}

	method, ok := m["method"].(string)
	if !ok || method != "isPrime" {
		return Malformed
	}

	// TODO: handle big int
	number, ok := m["number"].(float64)
	if !ok {
		return Malformed
	}

	return isPrime(int(number))
}

func isPrime(n int) ResponseType {
	if n <= 1 {
		return NotPrime
	}

	var i int = 2
	for ; (i * i) <= n; i += 1 {
		if n%i == 0 {
			return NotPrime
		}
	}

	return Prime
}
