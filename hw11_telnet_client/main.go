package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//nolint:gocritic
func main() {
	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Minimum 2 args: host & port")
		os.Exit(1)
	}

	address := net.JoinHostPort(args[0], args[1])

	ctx, cancelFunction := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancelFunction()

	telnetClient := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)

	err := telnetClient.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connection failed: %v", err)
		os.Exit(1)
	}
	defer func() {
		_ = telnetClient.Close()
	}()

	fmt.Fprintf(os.Stderr, "Connected to %s\n", address)

	errChan := make(chan error, 2)

	go func() {
		errChan <- telnetClient.Send()
	}()

	go func() {
		errChan <- telnetClient.Receive()
	}()

	select {
	case <-ctx.Done():
		fmt.Fprintf(os.Stderr, "SIGINT received\n")
	case err := <-errChan:
		if err != nil {
			fmt.Fprintf(os.Stderr, "Telnet error: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Connection was closed by peer\n")
		}
	}
	fmt.Fprintln(os.Stderr, "Bye-bye")
}
