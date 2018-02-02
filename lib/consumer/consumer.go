package consumer

import (
	"github.com/PumpkinSeed/npc/lib/common"
	nsq "github.com/nsqio/go-nsq"
)

// Config collects configuration parameters
type Config struct {
	// NSQConfig config for the consumer
	NSQConfig *nsq.Config

	// NSQDAddress address of the NSQ daemon
	NSQDAddress string

	// NSQLookupdAddresses addresses of the NSQ Lookup daemons
	NSQLookupdAddresses []string

	// Concurrency amout of concurrent handlers of the consumer
	Concurrency int

	// Logger of the consumer
	Logger common.Logger

	// LogLevel set the log level of the consumer
	LogLevel nsq.LogLevel
}

// nsqConfig checks the NSQConfig, if nil then initialize it
func (c *Config) nsqConfig() *nsq.Config {
	if c.NSQConfig == nil {
		c.NSQConfig = nsq.NewConfig()
	}
	return c.NSQConfig
}

// New creates and configures new nsq.Consumer.
func New(cfg *Config, topic, channel string, handler nsq.Handler) (*nsq.Consumer, error) {
	// get the consumer for the given topic
	consumer, err := nsq.NewConsumer(topic, channel, cfg.nsqConfig())
	if err != nil {
		return nil, err
	}

	// setup the logger
	consumer.SetLogger(cfg.Logger, cfg.LogLevel)

	consumer.AddHandler(handler)

	// add concurrent handlers
	//consumer.AddConcurrentHandlers(handler, cfg.Concurrency)

	// based on the defined addresses connect to the NSQ cluster
	if addrs := cfg.NSQLookupdAddresses; addrs != nil {
		if err := consumer.ConnectToNSQLookupds(addrs); err != nil {
			return nil, err
		}
	} else {
		if err := consumer.ConnectToNSQD(cfg.NSQDAddress); err != nil {
			return nil, err
		}
	}

	return consumer, nil
}
