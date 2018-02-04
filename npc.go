package npc

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PumpkinSeed/npc/lib/common"
	"github.com/PumpkinSeed/npc/lib/consumer"
	"github.com/PumpkinSeed/npc/lib/producer"
	"github.com/PumpkinSeed/npc/lib/rpc"
)

// T the type of the npc
type T int

const (
	// Server type
	Server T = iota

	// Client type
	Client
)

// Main handler of the rpc framework
// server and client booleans determine the type of the handler
// p is config for producer, c is config for consumer
// reqTopic determine the request topic name
// logger stores the nsq inner logger
// err stores the errors occured in the setup and return at the end of the setup
// channel stores the name of the channel for the topics
// rspTopic define the response topic for the client
// app stores the server related AppServer
// interruptor stores the function called at the end of the server to handle custom interruption
// customInterruptor boolean true if custum interruption setted up
type Main struct {
	server bool
	client bool

	// Common data for both handler
	p        *producer.Config
	c        *consumer.Config
	reqTopic string
	logger   common.Logger
	err      error
	channel  string

	// Client related data
	rspTopic string

	// Server releated data
	app               rpc.AppServer
	interruptor       func()
	customInterruptor bool
}

// New creates a new instance of the Main handler based on the type
func New(t T) *Main {
	m := new(Main)
	switch t {
	case Server:
		m.server = true
	case Client:
		m.client = true
	}

	return m
}

// Init method do the initial setup for both handler
// if the configs are nil values it will save an error for later
func (m *Main) Init(p *producer.Config, c *consumer.Config, rt string, channel string, logger common.Logger) *Main {
	m.p = p
	m.c = c
	m.reqTopic = rt
	m.logger = logger
	m.channel = channel

	if m.p == nil {
		m.err = errors.New("empty producer config")
		return m
	}
	if m.c == nil {
		m.err = errors.New("empty consumer config")
		return m
	}

	return m
}

/*
	Server related methods
*/

// Server do the setup for the server kind of handler
// it waits an AppServer as parameter
// return the error occured in the previous steps
func (m *Main) Server(app rpc.AppServer) (*Main, error) {
	if m.client {
		return nil, errors.New("client can't act as a server")
	}

	m.app = app

	return m, m.err
}

// Listen starts the server and listen on the request topic
// interruptor stops it, if it's triggered
func (m *Main) Listen() error {
	if m.client {
		return errors.New("client can't act as a server")
	}

	var err error

	p, err := producer.New(m.p)

	// rpc server: accepts request, calls application, sends response
	ctx, cancel := context.WithCancel(context.Background())
	rpcServer := rpc.NewServer(ctx, m.app, p)

	c, err := consumer.New(m.c, m.reqTopic, m.channel, rpcServer)
	if err != nil {
		cancel()
		return err
	}

	// clean exit
	defer p.Stop() // 3. stop response producer
	defer cancel() // 2. cancel any pending operation (returns unfinished messages to nsq)
	defer c.Stop() // 1. stop accepting new requestser.Stop() // 1. stop accepting new requests

	if m.customInterruptor {
		m.interruptor()
	} else {
		DefaultInterupt()
	}

	return nil
}

// SetInterruptor setup a custom interruptor
func (m *Main) SetInterruptor(i func()) {
	m.customInterruptor = true
	m.interruptor = i
}

/*
	Client related methods
*/

// Client do the setup for client kind of handler
// it waits for a response topic name as parameter
func (m *Main) Client(rt string) (*Main, error) {
	if m.server {
		return nil, errors.New("server can't act as a client")
	}

	m.rspTopic = rt

	return m, m.err
}

// Publish a message to the nsq and waits for the server response
// through the response topic
// the first parameter is the resource what it want to reach from the
// AppServer and the second is the exact message
func (m *Main) Publish(typ string, msg []byte) ([]byte, error) {
	if m.server {
		return nil, errors.New("server can't act as a client")
	}

	var err error

	p, err := producer.New(m.p)

	// rpc client: sends requests, waits and accepts responses
	//             provides interface for application
	rpcClient := rpc.NewClient(p, m.reqTopic, m.rspTopic)

	c, err := consumer.New(m.c, m.rspTopic, m.channel, rpcClient)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	// clean exit
	defer p.Stop() // 3. stop producing new requests
	defer cancel() // 2. cancel any pending (waiting for responses)
	defer c.Stop() // 1. stop listening for responses

	rspBody, rspErr, err := rpcClient.Call(ctx, typ, msg)
	if err != nil {
		return nil, err
	}
	if rspErr != "" {
		return nil, errors.New(rspErr)
	}

	return rspBody, nil
}

// DefaultInterupt is the default one and waits for a Ctrl+C
func DefaultInterupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}
