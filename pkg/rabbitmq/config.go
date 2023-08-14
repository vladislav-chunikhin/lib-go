package rabbitmq

import "time"

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
