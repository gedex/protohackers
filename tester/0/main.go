package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
}

func startServer(ctx context.Context) error {
	bin := os.Getenv("SERVER_BIN")
	if bin == "" {
		return errors.New("missing SERVER_BIN env")
	}

	args := strings.Fields(os.Getenv("SERVER_ARGS"))
	if len(args) == 0 {
		return errors.New("missing SERVER_ARGS env")
	}

	server := exec.CommandContext(ctx, bin, args...)
	if err := server.Start(); err != nil {
		return err
	}

	// Wait a bit for the server to start.
	time.Sleep(1 * time.Second)

	return nil
}
