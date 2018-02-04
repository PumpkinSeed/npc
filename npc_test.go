package npc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/PumpkinSeed/npc/lib/common"
	"github.com/PumpkinSeed/npc/lib/consumer"
	"github.com/PumpkinSeed/npc/lib/producer"
	nsq "github.com/nsqio/go-nsq"
)

var localNSQd = "127.0.0.1:4150"
var wg sync.WaitGroup
var l = common.SingleLogger{}

func TestAll(t *testing.T) {
	wg.Add(1)
	go s(t)

	wg.Wait()
}

func c(t *testing.T) {
	pConf := &producer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}

	// create consumer arround rpcClient
	cConf := &consumer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}

	m, err := New(Client).
		Init(pConf, cConf, "request", "server", l).
		Client("response")

	if err != nil {
		panic(err)
	}

	resp, err := m.Publish("Add", []byte("test"))
	if err != nil {
		panic(err)
	}

	fmt.Println(string(resp))

	wg.Done()
}

func s(t *testing.T) {
	pConf := &producer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}

	cConf := &consumer.Config{
		NSQConfig:   nsq.NewConfig(),
		NSQDAddress: localNSQd,
		Logger:      l,
		LogLevel:    nsq.LogLevelInfo,
	}

	m, err := New(Server).
		Init(pConf, cConf, "request", "server", common.SingleLogger{}).
		Server(&app{})

	if err != nil {
		t.Error(err)
		return
	}

	m.SetInterruptor(
		func() {

		},
	)

	err = m.Listen()
	if err != nil {
		t.Error(err)
		return
	}
}

type app struct{}

func (a *app) Serve(ctx context.Context, method string, reqBuf []byte) ([]byte, error) {
	switch method {
	case "Add":
		rsp := struct {
			Z int
		}{
			Z: 12,
		}
		rspBuf, err := json.Marshal(rsp)
		if err != nil {
			return nil, err
		}
		return rspBuf, nil
	default:
		return nil, nil
	}
}
