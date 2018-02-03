// +build server

package main

import (
	"context"
	"encoding/json"

	"github.com/PumpkinSeed/npc"
	"github.com/PumpkinSeed/npc/lib/common"
	"github.com/PumpkinSeed/npc/lib/consumer"
	"github.com/PumpkinSeed/npc/lib/producer"
	nsq "github.com/nsqio/go-nsq"
)

var localNSQd = "127.0.0.1:4150"

func main() {
	l := common.SingleLogger{}

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

	m, err := npc.New(npc.Server).
		Init(pConf, cConf, "request", "server", common.SingleLogger{}).
		Server(&app{})

	if err != nil {
		panic(err)
		return
	}

	m.Listen()
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
