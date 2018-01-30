package consumer

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	nsq "github.com/nsqio/go-nsq"
)

type Config struct {
	Log     bool
	Channel string
	Topic   string
	Handler nsq.Handler
}

func NewConfig(topic, ch string, log bool, handler nsq.Handler) *Config {
	return &Config{
		Topic:   topic,
		Channel: ch,
		Log:     log,
		Handler: handler,
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

// DefaultMessageHandler adheres to the nsq.Handler interface.
// This allows us to define our own custome handlers for
// our messages. Think of these handlers much like you would
// an http handler.
type DefaultMessageHandler struct{}

// HandleMessage is the only requirement needed to fulfill the
// nsq.Handler interface. This where you'll write your message
// handling logic.
func (h *DefaultMessageHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// returning an error results in the message being re-enqueued
		// a REQ is sent to nsqd
		return errors.New("body is blank re-enqueue message")
	}

	// Let's log our message!
	fmt.Println(string(m.Body))

	// Returning nil signals to the consumer that the message has
	// been handled with success. A FIN is sent to nsqd
	return nil
}

func Start(c *Config) {
	// The default config settings provide a pretty good starting point for
	// our new consumer.
	config := nsq.NewConfig()

	// Create a NewConsumer with the name of our topic, the channel, and our config
	consumer, err := nsq.NewConsumer(c.Topic, c.Channel, config)
	if err != nil {
		log.Fatal(err) // If there is an error, halt the application
	}

	// Set the number of messages that can be in flight at any given time
	// you'll want to set this number as the default is only 1. This can be
	// a major concurrency knob that can change the performance of your application.
	consumer.ChangeMaxInFlight(200)

	if c.Log {
		consumer.SetLogger(
			&innerLogger{},
			nsq.LogLevelInfo,
		)
	} else {
		consumer.SetLogger(
			&noopLogger{},
			nsq.LogLevelError,
		)
	}

	// Injects our handler into the consumer. You'll define one handler
	// per consumer, but you can have as many concurrently running handlers
	// as specified by the second argument. If your MaxInFlight is less
	// than your number of concurrent handlers you'll  starve your workers
	// as there will never be enough in flight messages for your worker pool
	consumer.AddConcurrentHandlers(
		c.Handler,
		20,
	)

	// Our consumer will discover where topics are located by our three
	// nsqlookupd instances The application will periodically poll
	// these nqslookupd instances to discover new nodes or drop unhealthy
	// producers.
	/*nsqlds := []string{"localhost:4161", "localhost:4163", "localhost:4165"}
	if err := consumer.ConnectToNSQLookupds(nsqlds); err != nil {
		log.Fatal(err)
	}*/

	err = consumer.ConnectToNSQD("localhost:4150")
	if err != nil {
		log.Panic("Could not connect")
	}

	// Let's allow our queues to drain properly during shutdown.
	// We'll create a channel to listen for SIGINT (Ctrl+C) to signal
	// to our application to gracefully shutdown.
	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT)

	// This is our main loop. It will continue to read off of our nsq
	// channel until either the consumer dies or our application is signaled
	// to stop.
	for {
		select {
		case <-consumer.StopChan:
			return // uh oh consumer disconnected. Time to quit.
		case <-shutdown:
			// Synchronously drain the queue before falling out of main
			consumer.Stop()
		}
	}
}
