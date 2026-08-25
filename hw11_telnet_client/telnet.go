package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type MyClient struct {
	address    string
	timeout    time.Duration
	in         io.ReadCloser
	out        io.Writer
	connection net.Conn
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &MyClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

func (client *MyClient) Connect() error {
	dialer := &net.Dialer{
		Timeout: client.timeout,
	}
	connection, err := dialer.DialContext(context.Background(), "tcp", client.address)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	client.connection = connection
	return nil
}

func (client *MyClient) Close() error {
	if client.connection != nil {
		err := client.connection.Close()
		if err != nil {
			return fmt.Errorf("disconnection error: %w", err)
		}
	}
	return nil
}

func (client *MyClient) Send() error {
	if client.connection == nil {
		return errors.New("telnet client is not connected")
	}
	_, err := io.Copy(client.connection, client.in)
	if err != nil {
		return fmt.Errorf("sending error: %w", err)
	}
	if tcpConn, ok := client.connection.(*net.TCPConn); ok {
		_ = tcpConn.CloseWrite()
	}
	return nil
}

func (client *MyClient) Receive() error {
	if client.connection == nil {
		return errors.New("telnet client is not connected")
	}
	_, err := io.Copy(client.out, client.connection)
	if err != nil {
		return fmt.Errorf("receiving error: %w", err)
	}
	return nil
}
