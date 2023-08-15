package redis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/vladislav-chunikhin/lib-go/pkg/logger"
)

var ErrCacheMiss = errors.New("cache: key not found")

type Config struct {
	Username     string        `yaml:"username"`
	Password     string        `yaml:"password"`
	Address      string        `yaml:"address"`
	PoolSize     int           `yaml:"poolSize"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
	DialTimeout  time.Duration `yaml:"dialTimeout"`
}

type Client struct {
	client *redis.Client
}

func NewClient(cfg *Config) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Username: cfg.Username,
		Password: cfg.Password,

		Addr: cfg.Address,

		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,

		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	if status := client.Ping(ctx); status != nil && status.Err() != nil {
		return nil, status.Err()
	}

	return &Client{client: client}, nil
}

func (c *Client) Get(ctx context.Context, key string, ptrValue any) error {
	b, err := c.client.Get(ctx, key).Bytes()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCacheMiss
		}
		return fmt.Errorf("cache problem while retrieving value from redis, key `%s`: %w", key, err)
	}

	return deserialize(b, ptrValue)
}

func (c *Client) Set(ctx context.Context, key string, ptrValue any, expires time.Duration) error {
	b, err := serialize(ptrValue)
	if err != nil {
		return fmt.Errorf("cache problem while serializing value before cache set, key `%s`: %w", key, err)
	}

	if err = c.client.Set(ctx, key, b, expires).Err(); err != nil {
		return fmt.Errorf("cache problem while set value to redis, key `%s`: %w", key, err)
	}

	return nil
}

func (c *Client) Close(logger logger.Logger) error {
	if err := c.client.Close(); err != nil {
		return err
	}
	logger.Debugf("redis disconnecting...")
	return nil
}

func deserialize(byt []byte, ptr any) error {

	if data, ok := ptr.(*[]byte); ok {
		*data = byt
		return nil
	}

	b := bytes.NewBuffer(byt)
	decoder := json.NewDecoder(b)
	if err := decoder.Decode(ptr); err != nil {
		return err
	}

	return nil
}

func serialize(value any) ([]byte, error) {

	if data, ok := value.([]byte); ok {
		return data, nil
	}

	var b bytes.Buffer
	encoder := json.NewEncoder(&b)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
