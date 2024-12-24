package rabbit

import (
	"context"
	"fmt"

	"github.com/geffersonFerraz/frigate-s3-telegram/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ interface {
	Publish(ctx context.Context, message []byte) error
	Consume(handler func([]byte) error) error
	Close() error
}

type rabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewRabbitMQ() (RabbitMQ, error) {
	cfg := config.New()

	conn, err := amqp.Dial(cfg.RabbitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		cfg.RabbitQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &rabbitMQ{
		conn:    conn,
		channel: ch,
		queue:   q,
	}, nil
}

func (r *rabbitMQ) Publish(ctx context.Context, message []byte) error {
	return r.channel.PublishWithContext(
		ctx,
		"",           // exchange
		r.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         message,
		},
	)
}

func (r *rabbitMQ) Consume(handler func([]byte) error) error {
	msgs, err := r.channel.Consume(
		r.queue.Name, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for msg := range msgs {
			err := handler(msg.Body)
			if err != nil {
				msg.Nack(false, true)
				continue
			}
			msg.Ack(false)
		}
	}()

	return nil
}

func (r *rabbitMQ) Close() error {
	if err := r.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := r.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}
