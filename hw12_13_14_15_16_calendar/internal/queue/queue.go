package queue

import "context"

type Producer interface {
	Publish(ctx context.Context, body []byte) error
	Close() error
}

type Consumer interface {
	Consume(ctx context.Context, handler func(body []byte) error) error
	Close() error
}
