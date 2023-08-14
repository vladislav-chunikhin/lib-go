package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/vladislav-chunikhin/lib-go/pkg/logger"
)

type Producer struct {
	conn       *amqp.Connection
	cfg        *Config
	logger     logger.Logger
	producerCh *amqp.Channel
}

func NewProducer(cfg *Config, logger logger.Logger) (*Producer, error) {
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

	return &Producer{
		conn:       conn,
		cfg:        cfg,
		logger:     logger,
		producerCh: producerCh,
	}, nil
}

func (c *Producer) DeclareQueue(queueName string) error {
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

func (c *Producer) Publish(ctx context.Context, message []byte, queueName string) error {
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
			Body:        message,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Producer) Close() error {
	if err := c.producerCh.Close(); err != nil {
		return err
	}

	if err := c.conn.Close(); err != nil {
		return err
	}

	c.logger.Debugf("rabbitmq disconnecting...")
	return nil
}
