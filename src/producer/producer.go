package producer

import (
	"fmt"
	"log"

	"github.com/nsqio/go-nsq"
)

// NoopNSQLogger allows us to pipe NSQ logs to dev/null
// The default NSQ logger is great for debugging, but did
// not fit our normally well structured JSON logs. Luckily
// NSQ provides a simple interface for injecting your own
// logger.
type NoopNSQLogger struct{}

// Output allows us to implement the nsq.Logger interface
func (l *NoopNSQLogger) Output(calldepth int, s string) error {
	fmt.Printf("--- producer --- > %s\n", s)
	return nil
}

func Write(message []byte) {
	config := nsq.NewConfig()
	producer, _ := nsq.NewProducer("localhost:4150", config)
	producer.SetLogger(
		&NoopNSQLogger{},
		nsq.LogLevelInfo,
	)

	err := producer.Publish("clicks", message)
	if err != nil {
		log.Panic("Could not connect")
	}

	producer.Stop()
}
