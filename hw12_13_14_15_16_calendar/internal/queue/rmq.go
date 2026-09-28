package queue

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RmqClient struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	queue      *amqp.Queue
}

func NewProducer(url, queueName string) (Producer, error) {
	client, err := newRmqClient(url, queueName)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewConsumer(url, queueName string) (Consumer, error) {
	client, err := newRmqClient(url, queueName)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func newRmqClient(url, queueName string) (*RmqClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable (сохраняется при перезапуск брокера)
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	return &RmqClient{
		connection: conn,
		channel:    ch,
		queue:      &q,
	}, nil
}

func (r *RmqClient) Publish(ctx context.Context, body []byte) error {
	return r.channel.PublishWithContext(
		ctx,
		"",           // exchange
		r.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (r *RmqClient) Consume(ctx context.Context, handler func([]byte) error) error {
	if err := r.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set qos: %w", err)
	}

	msgs, err := r.channel.Consume(
		r.queue.Name, // queue
		"",           // consumer
		false,        // auto-ack = false (ручное подтверждение получения)
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return fmt.Errorf("error of consumer creation: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case delivery, ok := <-msgs:
				if !ok {
					return
				}
				if err := handler(delivery.Body); err != nil {
					// return to queue
					_ = delivery.Nack(false, true)
				} else {
					// successfully received
					_ = delivery.Ack(false)
				}
			}
		}
	}()

	return nil
}

func (r *RmqClient) Close() error {
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.connection != nil {
		return r.connection.Close()
	}
	return nil
}
