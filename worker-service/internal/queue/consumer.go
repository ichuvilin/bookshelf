package queue

import (
	"errors"
	"fmt"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

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

type Consumer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	handlers map[string]HandlerFunc

	mu           sync.Mutex
	wg           sync.WaitGroup
	stopOnce     sync.Once
	consumerTags map[string]string
	started      bool
	stopErr      error
}

func NewConsumer(url string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open RabbitMQ channel: %w", err)
	}

	return &Consumer{
		conn:         conn,
		channel:      ch,
		handlers:     make(map[string]HandlerFunc),
		consumerTags: make(map[string]string),
	}, nil
}

func (c *Consumer) RegisterHandler(queue string, handler HandlerFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		log.Printf("cannot register handler for %q after Start", queue)
		return
	}

	c.handlers[queue] = handler
}

func (c *Consumer) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return errors.New("consumer already started")
	}

	for queue, handler := range c.handlers {
		if _, err := c.channel.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("declare queue %q: %w", queue, err)
		}

		if err := c.channel.Qos(1, 0, false); err != nil {
			return fmt.Errorf("set QoS for queue %q: %w", queue, err)
		}

		tag := fmt.Sprintf("worker-%s", queue)

		msgs, err := c.channel.Consume(
			queue,
			tag,
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("consume queue %q: %w", queue, err)
		}

		c.consumerTags[queue] = tag
		c.wg.Add(1)

		go c.consume(queue, handler, msgs)
	}

	c.started = true
	return nil
}

func (c *Consumer) consume(
	queue string,
	handler HandlerFunc,
	msgs <-chan amqp.Delivery,
) {
	defer c.wg.Done()

	for msg := range msgs {
		log.Printf(
			"received message from queue %q: %s",
			queue,
			string(msg.Body),
		)

		err := handler(msg.Body)
		if err == nil {
			if ackErr := msg.Ack(false); ackErr != nil {
				log.Printf("failed to ack message from %q: %v", queue, ackErr)
			}
			continue
		}

		var temporaryErr *TemporaryError

		if errors.As(err, &temporaryErr) {
			log.Printf(
				"temporary error processing message from %q: %v",
				queue,
				err,
			)

			if nackErr := msg.Nack(false, true); nackErr != nil {
				log.Printf("failed to nack message from %q: %v", queue, nackErr)
			}
			continue
		}

		log.Printf(
			"permanent error processing message from %q: %v",
			queue,
			err,
		)

		if nackErr := msg.Nack(false, false); nackErr != nil {
			log.Printf("failed to nack message from %q: %v", queue, nackErr)
		}
	}
}

// Stop stops accepting new deliveries and waits for active consumers
// to finish processing deliveries already received.
func (c *Consumer) Stop() error {
	c.stopOnce.Do(func() {
		c.mu.Lock()

		if !c.started {
			c.mu.Unlock()
			c.stopErr = c.CloseConnections()
			return
		}

		log.Println("Cancelling RabbitMQ consumers...")

		for queue, tag := range c.consumerTags {
			if err := c.channel.Cancel(tag, false); err != nil {
				log.Printf(
					"failed to cancel consumer for queue %q: %v",
					queue,
					err,
				)
				if c.stopErr == nil {
					c.stopErr = fmt.Errorf(
						"cancel consumer for queue %q: %w",
						queue,
						err,
					)
				}
			}
		}

		c.mu.Unlock()

		log.Println("Waiting for active message handlers...")
		c.wg.Wait()

		log.Println("All message handlers stopped")

		if err := c.CloseConnections(); err != nil && c.stopErr == nil {
			c.stopErr = err
		}
	})

	return c.stopErr
}

// Close is kept for compatibility with existing callers.
func (c *Consumer) Close() error {
	return c.Stop()
}

// CloseConnections closes the channel and connection after consumers stop.
func (c *Consumer) CloseConnections() error {
	var firstErr error

	if c.channel != nil {
		if err := c.channel.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
			firstErr = fmt.Errorf("close RabbitMQ channel: %w", err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) && firstErr == nil {
			firstErr = fmt.Errorf("close RabbitMQ connection: %w", err)
		}
	}

	return firstErr
}

func (c *Consumer) HealthCheck() error {
	if c.conn.IsClosed() {
		return errors.New("connection closed")
	}

	if c.channel.IsClosed() {
		return errors.New("channel closed")
	}

	return nil
}
