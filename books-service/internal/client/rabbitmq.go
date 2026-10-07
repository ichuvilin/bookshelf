package client

import (
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	QueueImageCompress = "image_compress"
)

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
	}, err
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
