package rpc

import (
	"context"
	"testing"

	"github.com/PumpkinSeed/nsq-rpc/lib/common"
	"github.com/PumpkinSeed/nsq-rpc/lib/consumer"
	"github.com/PumpkinSeed/nsq-rpc/lib/producer"
	nsq "github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"
)

/*
	The following tests requires local nsqd
*/

var localNSQd = "127.0.0.1:4150"

func TestServer(t *testing.T) {
	l := common.SingleLogger{}

	pConf := &producer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}

	p, err := producer.New(pConf)
	assert.Nil(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	appServer := &server{}

	rpcServer := NewServer(ctx, appServer, p)

	cConf := &consumer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}
	c, err := consumer.New(cConf, "request", "server", rpcServer)
	assert.Nil(t, err)

	defer p.Stop()
	defer cancel()
	defer c.Stop()
}

type server struct{}

func (s *server) Serve(ctx context.Context, method string, reqBuf []byte) ([]byte, error) {
	return []byte("rdy"), nil
}
