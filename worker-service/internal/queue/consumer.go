package queue

import (
	"errors"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	handlers map[string]HandlerFunc
}

type HandlerFunc func(body []byte) error

type TemporaryError struct {
	Err error
}

func (e *TemporaryError) Error() string {
	return e.Err.Error()
}

func (e *TemporaryError) Unwrap() error {
	return e.Err
}

func NewTemporaryError(err error) error {
	return &TemporaryError{Err: err}
}

func NewConsumer(url string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &Consumer{
		conn:     conn,
		channel:  ch,
		handlers: make(map[string]HandlerFunc),
	}, nil
}

func (c *Consumer) RegisterHandler(queue string, handler HandlerFunc) {
	c.handlers[queue] = handler
}

func (c *Consumer) Start() error {
	for queue, handler := range c.handlers {
		if _, err := c.channel.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			nil); err != nil {
			return err
		}

		if err := c.channel.Qos(1, 0, false); err != nil {
			return err
		}

		msgs, err := c.channel.Consume(
			queue,
			"",
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return err
		}

		go c.consume(queue, handler, msgs)
	}

	return nil
}

func (c *Consumer) consume(
	queue string,
	handler HandlerFunc,
	msgs <-chan amqp.Delivery,
) {
	for msg := range msgs {
		log.Printf("received message from queue %q: %s", queue, string(msg.Body))

		err := handler(msg.Body)
		if err == nil {
			msg.Ack(false)
			continue
		}

		var temporaryErr *TemporaryError

		if errors.As(err, &temporaryErr) {
			log.Printf(
				"temporary error processing message from queue %q: %v",
				queue,
				err,
			)

			if nackErr := msg.Nack(false, true); nackErr != nil {
				log.Printf(
					"failed to nack message from queue %q: %v",
					queue,
					nackErr,
				)
			}

			continue
		}

		log.Printf(
			"permanent error processing message from queue %q: %v",
			queue,
			err,
		)

		if nackErr := msg.Nack(false, false); nackErr != nil {
			log.Printf(
				"failed to nack message from queue %q: %v",
				queue,
				nackErr,
			)
		}
	}
}

func (c *Consumer) Close() error {
	if err := c.channel.Close(); err != nil {
		_ = c.conn.Close()
		return err
	}

	return c.conn.Close()
}
