package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/vladislav-chunikhin/lib-go/pkg/logger"
)

type Config struct {
	URL      string         `yaml:"url"`
	Timeout  time.Duration  `yaml:"timeout"`
	Queue    QueueConfig    `yaml:"queue"`
	Consumer ConsumerConfig `yaml:"consumer"`
}

type QueueConfig struct {
	Durable    bool `yaml:"durable"`
	AutoDelete bool `yaml:"autoDelete"`
	Exclusive  bool `yaml:"exclusive"`
	NoWait     bool `yaml:"noWait"`
}

type ConsumerConfig struct {
	Name      string `yaml:"name"`
	AutoAck   bool   `yaml:"autoAck"`
	Exclusive bool   `yaml:"exclusive"`
	NoLocal   bool   `yaml:"noLocal"`
	NoWait    bool   `yaml:"noWait"`
}

type Client struct {
	conn   *amqp.Connection
	cfg    *Config
	logger logger.Logger

	producerCh *amqp.Channel
	consumerCh *amqp.Channel
}

func NewClient(cfg *Config, logger logger.Logger) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil cfg")
	}

	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to create amqp connection: %w", err)
	}

	var producerCh *amqp.Channel
	producerCh, err = conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create producer channel: %w", err)
	}

	var consumerCh *amqp.Channel
	consumerCh, err = conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer channel: %w", err)
	}

	return &Client{
		conn:       conn,
		cfg:        cfg,
		logger:     logger,
		producerCh: producerCh,
		consumerCh: consumerCh,
	}, nil
}

func (c *Client) DeclareQueue(queueName string) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		queueName,
		c.cfg.Queue.Durable,
		c.cfg.Queue.AutoDelete,
		c.cfg.Queue.Exclusive,
		c.cfg.Queue.NoWait,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Publish(ctx context.Context, message, queueName string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	err := c.producerCh.PublishWithContext(
		timeoutCtx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Consume(ctx context.Context, queueName string, handler func([]byte) error) error {
	msgs, err := c.consumerCh.Consume(
		queueName,
		c.cfg.Consumer.Name,
		c.cfg.Consumer.AutoAck,
		c.cfg.Consumer.Exclusive,
		c.cfg.Consumer.NoLocal,
		c.cfg.Consumer.NoWait,
		nil,
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			c.logger.Debugf("amqp consumer work stopped")
			return nil
		case msg := <-msgs:
			if err = handler(msg.Body); err != nil {
				msg.Ack(false)
				c.logger.Debugf("message was handled with problems: %v", err)
			} else {
				msg.Ack(true)
				c.logger.Debugf("message was handled successfully: %v", err)
			}
		}
	}
}

func (c *Client) Close() error {
	if err := c.producerCh.Close(); err != nil {
		return err
	}

	if err := c.consumerCh.Close(); err != nil {
		return err
	}

	if err := c.conn.Close(); err != nil {
		return err
	}

	c.logger.Debugf("rabbitmq disconnecting...")
	return nil
}
