package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/vladislav-chunikhin/lib-go/pkg/logger"
)

type Consumer struct {
	conn       *amqp.Connection
	cfg        *Config
	logger     logger.Logger
	consumerCh *amqp.Channel
}

func NewConsumer(cfg *Config, logger logger.Logger) (*Consumer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil cfg")
	}

	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to create amqp connection: %w", err)
	}

	var consumerCh *amqp.Channel
	consumerCh, err = conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer channel: %w", err)
	}

	return &Consumer{
		conn:       conn,
		cfg:        cfg,
		logger:     logger,
		consumerCh: consumerCh,
	}, nil
}

func (c *Consumer) DeclareQueue(queueName string) error {
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

func (c *Consumer) Consume(ctx context.Context, queueName string, handler func([]byte) error) error {
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

func (c *Consumer) Close() error {
	if err := c.consumerCh.Close(); err != nil {
		return err
	}

	if err := c.conn.Close(); err != nil {
		return err
	}

	c.logger.Debugf("rabbitmq disconnecting...")
	return nil
}
