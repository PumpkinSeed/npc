package npc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/PumpkinSeed/npc/lib/common"
	"github.com/PumpkinSeed/npc/lib/consumer"
	"github.com/PumpkinSeed/npc/lib/producer"
)

func TestNewServer(t *testing.T) {

	m := New(Server).
		Init(&producer.Config{}, &consumer.Config{}, "request", common.SingleLogger{}).
		Server(&app{}, "server")

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
