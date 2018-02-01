package rpc

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

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
	//var wg sync.WaitGroup

	//wg.Add(1)
	//go s(t, &wg)

	l := common.SingleLogger{}

	pConf := &producer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}

	p, err := producer.New(pConf)
	assert.Nil(t, err)

	rpcClient := NewClient(p, "request", "response")

	cConf := &consumer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}
	c, err := consumer.New(cConf, "response", "server", rpcClient)
	assert.Nil(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	// clean exit
	defer p.Stop() // 3. stop producing new requests
	defer cancel() // 2. cancel any pending (waiting for responses)
	defer c.Stop() // 1. stop listening for responses

	rsp, rspStr, err := rpcClient.Call(ctx, "test", []byte("test"))
	assert.Nil(t, err)

	fmt.Println(rsp)
	fmt.Println(rspStr)

	//wg.Done()
}

func s(t *testing.T, wg *sync.WaitGroup) {
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

	wg.Wait()
}

type server struct{}

func (s *server) Serve(ctx context.Context, method string, reqBuf []byte) ([]byte, error) {
	return []byte("rdy"), nil
}
