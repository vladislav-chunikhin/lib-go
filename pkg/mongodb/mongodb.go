package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/vladislav-chunikhin/lib-go/pkg/logger"
)

type Config struct {
	URL            string        `yaml:"url"`
	User           string        `yaml:"user"`
	Password       string        `yaml:"password"`
	Timeout        time.Duration `yaml:"timeout"`
	ConnectTimeout time.Duration `yaml:"connectTimeout"`
	SocketTimeout  time.Duration `yaml:"socketTimeout"`
	MaxPoolSize    uint64        `yaml:"maxPoolSize"`
}

type Client struct {
	mongo *mongo.Client
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil cfg")
	}

	client, err := mongo.Connect(
		ctx,
		options.Client().SetAuth(options.Credential{
			Username: cfg.User,
			Password: cfg.Password,
		}),
		options.Client().ApplyURI(cfg.URL),
		options.Client().SetTimeout(cfg.Timeout),
		options.Client().SetConnectTimeout(cfg.ConnectTimeout),
		options.Client().SetSocketTimeout(cfg.SocketTimeout),
		options.Client().SetMaxPoolSize(cfg.MaxPoolSize),
	)

	if err != nil {
		return nil, err
	}

	return &Client{mongo: client}, nil
}

func (c *Client) GetCollection(db, collection string) *mongo.Collection {
	return c.mongo.Database(db).Collection(collection)
}

func (c *Client) Ping(ctx context.Context) error {
	if err := c.mongo.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}
	return nil
}

func (c *Client) Close(ctx context.Context, logger logger.Logger) error {
	if err := c.mongo.Disconnect(ctx); err != nil {
		return err
	}

	logger.Debugf("mongodb disconnecting...")
	return nil
}
