package npc

import (
	"errors"

	"github.com/PumpkinSeed/npc/lib/common"
	"github.com/PumpkinSeed/npc/lib/consumer"
	"github.com/PumpkinSeed/npc/lib/producer"
	"github.com/PumpkinSeed/npc/lib/rpc"
	nsq "github.com/nsqio/go-nsq"
)

// T the type of the npc
type T int

const (
	// Server type
	Server T = iota

	// Client type
	Client
)

// Main handler
type Main struct {
	server bool
	client bool

	// Common data for both handler
	p        *producer.Config
	producer *nsq.Producer
	c        *consumer.Config
	consumer *nsq.Consumer
	reqTopic string
	logger   common.Logger
	err      error

	// Client related data
	rspTopic  string
	resource  string
	rpcClient *rpc.Client

	// Server releated data
	app       rpc.AppServer
	channel   string
	rpcServer *rpc.Server
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

func (m *Main) Init(p *producer.Config, c *consumer.Config, rt string, logger common.Logger) *Main {
	m.p = p
	m.c = c
	m.reqTopic = rt
	m.logger = logger

	return m
}

func (m *Main) SetupNSQ() {
	var err error

	m.producer, err = producer.New(m.p)
	if err != nil {
		m.err = err
	}

	var handler nsq.Handler
	if m.server && m.rpcServer != nil {
		handler = m.rpcServer
	} else {
		m.err = errors.New("")
	}
	if m.client && m.rpcClient != nil {
		handler = m.rpcClient
	}

	m.consumer, err = consumer.New(m.c, m.reqTopic, m.channel, handler)
	if err != nil {
		m.err = err
	}
}

/*
	Server related methods
*/

func (m *Main) Server(app rpc.AppServer, channel string) *Main {
	m.app = app
	m.channel = channel

	return m
}

func (m *Main) Listen() error {

	return nil
}

/*
	Client related methods
*/

func (m *Main) Client(rt string, resource string) *Main {
	m.rspTopic = rt
	m.resource = resource

	return m
}
