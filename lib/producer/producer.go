package producer

import (
	"github.com/PumpkinSeed/nsq-rpc/lib/common"
	nsq "github.com/nsqio/go-nsq"
)

// Config collects configuration parameters
type Config struct {
	// NSQConfig config for the consumer
	NSQConfig *nsq.Config

	// NSQDAddress address of the NSQ daemon
	NSQDAddress string

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

// New creates nsq.Producer from Config.
func New(cfg *Config) (*nsq.Producer, error) {
	// get a producer for the nsq daemon
	producer, err := nsq.NewProducer(cfg.NSQDAddress, cfg.nsqConfig())
	if err != nil {
		return nil, err
	}

	// setup the logger
	producer.SetLogger(cfg.Logger, cfg.LogLevel)

	return producer, nil
}
