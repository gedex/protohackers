package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strconv"
	"sync"
	"testing"
	"time"
)

type ResponseType int

const (
	Malformed ResponseType = -1
	NotPrime               = 0
	Prime                  = 1
)

func getTestCaseFile(t *testing.T, testCase int, suffix string) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed get working directory: %s", err)
	}

	return path.Join(wd, fmt.Sprintf("./testdata/%d.%s", testCase, suffix))
}

func getTestCase(t *testing.T, testCase int) (requests []byte, expects []ResponseType) {
	in := getTestCaseFile(t, testCase, "in")
	requests, err := os.ReadFile(in)
	if err != nil {
		t.Fatalf("failed to read test file %s", in)
	}

	out := getTestCaseFile(t, testCase, "out")
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read test file %s", out)
	}
	lines := bytes.Split(b, []byte("\n"))

	// Convert 1, 0, -1 to ResponseType.
	for i, line := range lines {
		if len(line) == 0 {
			continue
		}

		o, err := strconv.ParseInt(string(line), 10, 0)
		if err != nil {
			t.Fatalf("failed to parse %s in %s:%d", string(line), out, i+1)
		}

		expects = append(expects, ResponseType(o))
	}

	return requests, expects
}

func respType(resp []byte) ResponseType {
	m := map[string]any{}
	err := json.Unmarshal(resp, &m)
	if err != nil {
		return Malformed
	}

	prime, ok := m["prime"].(bool)
	if !ok {
		return Malformed
	}

	if prime {
		return Prime
	}
	return NotPrime
}

func test(t *testing.T, conn net.Conn, testCase int) {
	requests, expects := getTestCase(t, testCase)
	if _, err := conn.Write(requests); err != nil {
		t.Fatalf("failed to write %d.in to conn: %s", testCase, err)
	}

	b, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("failed to read from conn: %s", err)
	}
	b = bytes.TrimRight(b, "\n")

	resp := bytes.Split(b, []byte("\n"))
	for i, r := range resp {
		actual := respType(r)
		if actual != expects[i] {
			t.Fatalf(
				"%s:%d, resp %s expect %v got %v",
				getTestCaseFile(t, testCase, "out"),
				i+1,
				r,
				expects[i],
				actual,
			)
		}
	}
}

func TestSimple1(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	if err := startServer(ctx); err != nil {
		t.Fatalf("failed to start server: %s", err)
	}
	defer cancel()

	conn, err := net.DialTimeout(
		"tcp",
		os.Getenv("SERVER_ARGS"),
		time.Second*2,
	)
	if err != nil {
		t.Fatalf("couldn't connect to the server: %s", err)
	}

	test(t, conn, 1)
	conn.Close()
}

func TestMultiClient2To6(t *testing.T) {
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	if err := startServer(ctx); err != nil {
		t.Fatalf("failed to start server: %s", err)
	}
	defer cancel()

	for i := 2; i <= 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := net.DialTimeout(
				"tcp",
				os.Getenv("SERVER_ARGS"),
				time.Second*2,
			)
			if err != nil {
				t.Fatalf("couldn't connect to the server: %s", err)
			}
			defer conn.Close()

			test(t, conn, i)
		}()
	}

	wg.Wait()
}

func TestMalformed7To23(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	if err := startServer(ctx); err != nil {
		t.Fatalf("failed to start server: %s", err)
	}
	defer cancel()

	for i := 7; i <= 23; i++ {
		conn, err := net.DialTimeout(
			"tcp",
			os.Getenv("SERVER_ARGS"),
			time.Second*2,
		)
		if err != nil {
			t.Fatalf("couldn't connect to the server: %s", err)
		}

		test(t, conn, i)
		conn.Close()
	}
}

func testWellFormed(t *testing.T, testCase int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	if err := startServer(ctx); err != nil {
		t.Fatalf("failed to start server: %s", err)
	}
	defer cancel()

	conn, err := net.DialTimeout(
		"tcp",
		os.Getenv("SERVER_ARGS"),
		time.Second*2,
	)
	if err != nil {
		t.Fatalf("couldn't connect to the server: %s", err)
	}
	test(t, conn, testCase)
	conn.Close()
}

func TestWellFormed24To31(t *testing.T) {
	for i := 24; i < 31; i++ {
		testWellFormed(t, i)
	}
}
