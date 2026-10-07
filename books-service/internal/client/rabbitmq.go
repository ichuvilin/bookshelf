package client

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	QueueImageCompress = "image_compress"
)

type ImageCompressMessage struct {
	BookID       string `json:"book_id"`
	CoverID      string `json:"cover_id"`
	OriginalPath string `json:"original_path"`
}

type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQClient{
		conn:    conn,
		channel: ch,
	}, nil
}

func (c *RabbitMQClient) DeclareQueue(name string) error {
	_, err := c.channel.QueueDeclare(name, true, false, false, false, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *RabbitMQClient) Close() error {
	err := c.channel.Close()
	if err != nil {
		return err
	}
	err = c.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *RabbitMQClient) HealthCheck() error {
	isClosed := c.conn.IsClosed()
	if isClosed {
		return errors.New("connection closed")
	}

	isClosed = c.channel.IsClosed()
	if isClosed {
		return errors.New("channel closed")
	}
	return nil
}

func (c *RabbitMQClient) publish(ctx context.Context, queue string, body []byte) error {
	if err := c.channel.PublishWithContext(ctx, "", queue, true, true, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Time{},
		Body:         body,
	}); err != nil {
		return err
	}
	return nil
}

func (c *RabbitMQClient) PublishImageCompress(ctx context.Context, msg ImageCompressMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	log.Printf("Published image compress message for book %s\n", msg.BookID)
	return c.publish(ctx, QueueImageCompress, data)
}
