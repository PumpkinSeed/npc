package main

import (
	"github.com/PumpkinSeed/durable-nsq-example/src/consumer"
	"github.com/PumpkinSeed/durable-nsq-example/src/producer"
)

func main() {
	go func() {
		for i := 0; i < 10; i++ {
			producer.Write([]byte("test"))
		}
	}()
	consumer.Start()
}
