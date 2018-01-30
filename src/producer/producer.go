package producer

import (
	"log"

	"github.com/nsqio/go-nsq"
)

type Config struct {
	Log   bool
	Topic string
}

func NewConfig(topic string, log bool) *Config {
	return &Config{
		Topic: topic,
		Log:   log,
	}
}

type innerLogger struct{}

func (l *innerLogger) Output(calldepth int, s string) error {
	return log.Output(calldepth, s)
}

// noopLogger allows us to pipe NSQ logs to dev/null
// The default NSQ logger is great for debugging, but did
// not fit our normally well structured JSON logs. Luckily
// NSQ provides a simple interface for injecting your own
// logger.
type noopLogger struct{}

// Output allows us to implement the nsq.Logger interface
func (l *noopLogger) Output(calldepth int, s string) error {
	return nil
}

func Write(message []byte, c *Config) {
	config := nsq.NewConfig()
	producer, _ := nsq.NewProducer("localhost:4150", config)

	if c.Log {
		producer.SetLogger(
			&innerLogger{},
			nsq.LogLevelInfo,
		)
	} else {
		producer.SetLogger(
			&noopLogger{},
			nsq.LogLevelError,
		)
	}

	err := producer.Publish(c.Topic, message)
	if err != nil {
		log.Panic("Could not connect")
	}

	producer.Stop()
}
