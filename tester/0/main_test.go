package main

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"path"
	"sync"
	"testing"
	"time"
)

var testDataFiles = []string{
	"./testdata/hello",
	"./testdata/hello_world",
	"./testdata/foobar",
	"./testdata/MIT-LICENSE",
	"./testdata/bin",
}

var testCases [][]byte
var testCasesInitialized = false

func setup(t *testing.T) {
	if testCasesInitialized {
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed get working directory: %s", err)
	}
	for _, f := range testDataFiles {
		fp := path.Join(wd, f)
		b, err := os.ReadFile(fp)
		if err != nil {
			t.Fatalf("failed to read test file %s", fp)
		}
		testCases = append(testCases, b)
	}

	testCasesInitialized = true
}

func testConn(t *testing.T, conn net.Conn, tc []byte) {
	if _, err := conn.Write(tc); err != nil {
		t.Fatalf("failed to write: %s", err)
	}

	actual := make([]byte, len(tc))
	if _, err := io.ReadAtLeast(conn, actual, len(tc)); err != nil {
		t.Fatalf("failed to read: %s", err)
	}

	if !bytes.Equal(tc, actual) {
		t.Fatalf("expect %v got %v", tc, actual)
	}
}

func TestOneClientAtATime(t *testing.T) {
	setup(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	if err := startServer(ctx); err != nil {
		t.Fatalf("failed to start server: %s", err)
	}
	defer cancel()

	for _, tc := range testCases {
		conn, err := net.DialTimeout("tcp", os.Getenv("SERVER_ARGS"), time.Second*2)
		if err != nil {
			t.Fatalf("couldn't connect to the server: %s", err)
		}

		testConn(t, conn, tc)
		conn.Close()
	}
}

func TestConcurrentClients(t *testing.T) {
	setup(t)
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	if err := startServer(ctx); err != nil {
		t.Fatalf("failed to start server: %s", err)
	}
	defer cancel()

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := net.DialTimeout("tcp", os.Getenv("SERVER_ARGS"), time.Second*2)
			if err != nil {
				t.Fatalf("couldn't connect to the server: %s", err)
			}
			defer conn.Close()

			for _, tc := range testCases {
				testConn(t, conn, tc)
			}
		}()
	}

	wg.Wait()
}
