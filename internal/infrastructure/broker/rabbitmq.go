package broker

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// RabbitMQ is the connection type
type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	log  *zap.Logger
	mu   sync.Mutex
	done chan struct{}
}

// RabbitURL is a tagged type to avoid string ambiguity in Wire
type RabbitURL string

func NewRabbitMQ(url RabbitURL, log *zap.Logger) (*RabbitMQ, error) {
	conn, err := amqp.Dial(string(url))
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	// Declare Topic Exchange (standar routing dinamis)
	if err := ch.ExchangeDeclare("app.events", "topic", true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	return &RabbitMQ{conn: conn, ch: ch, log: log, done: make(chan struct{})}, nil
}

func (r *RabbitMQ) Publish(ctx context.Context, routingKey string, payload []byte) error {
	return r.ch.PublishWithContext(ctx, "app.events", routingKey, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         payload,
			DeliveryMode: amqp.Persistent, // Persist ke disk jika broker restart
		})
}

func (r *RabbitMQ) Subscribe(topics []string, handler func(ctx context.Context, topic string, payload []byte) error) error {
	// Queue exclusive & auto-delete (tiap service/instance punya queue sendiri)
	q, err := r.ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return err
	}

	for _, t := range topics {
		if err := r.ch.QueueBind(q.Name, t, "app.events", false, nil); err != nil {
			return err
		}
	}

	msgs, err := r.ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			ctx := context.Background() // bisa inject trace/request ID middleware
			if err := handler(ctx, d.RoutingKey, d.Body); err != nil {
				r.log.Error("handler failed, dropping message (no requeue)",
					zap.String("topic", d.RoutingKey), zap.Error(err))
				d.Nack(false, false) // don't requeue - avoid infinite loop
			} else {
				d.Ack(false)
			}
		}
	}()
	return nil
}

func (r *RabbitMQ) Start(ctx context.Context) error { r.log.Info("rabbitmq ready"); return nil }
func (r *RabbitMQ) Stop() error {
	close(r.done)
	r.ch.Close()
	return r.conn.Close()
}
