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

var wg sync.WaitGroup

func TestServer(t *testing.T) {

	log("#-TestServer: test began")
	wg.Add(1)
	go s(t, &wg)

	log("#-TestServer: server started")
	l := common.SingleLogger{}

	pConf := &producer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}

	p, err := producer.New(pConf)
	assert.Nil(t, err)
	log("#-TestServer: client producer initialized")

	rpcClient := NewClient(p, "request", "response")
	log("#-TestServer: rpc client initialized")

	cConf := &consumer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}
	c, err := consumer.New(cConf, "response", "server", rpcClient)
	assert.Nil(t, err)
	log("#-TestServer: client consumer initialized")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// clean exit
	defer p.Stop() // 3. stop producing new requests
	defer cancel() // 2. cancel any pending (waiting for responses)
	defer c.Stop() // 1. stop listening for responses

	log("#-TestServer: rpc call")
	rsp, rspStr, err := rpcClient.Call(ctx, "test", []byte("test"))
	assert.Nil(t, err)
	log("#-TestServer: rpc happened")

	wg.Wait()
	fmt.Println(rsp)
	fmt.Println(rspStr)
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
	log("#-s: server producer initialized")

	ctx, cancel := context.WithCancel(context.Background())
	appServer := &server{}

	rpcServer := NewServer(ctx, appServer, p)
	log("#-s: rpc server initialized")

	cConf := &consumer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}
	c, err := consumer.New(cConf, "request", "server", rpcServer)
	assert.Nil(t, err)
	log("#-s: server consumer initialized")

	defer p.Stop()
	defer cancel()
	defer c.Stop()

	log("#-s: server waiting")
	defer log("#-s: server finished")

}

type server struct{}

func (s *server) Serve(ctx context.Context, method string, reqBuf []byte) ([]byte, error) {
	log("#-Serve: server message: rdy")
	return []byte("rdy"), nil
}
