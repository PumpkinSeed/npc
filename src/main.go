package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/PumpkinSeed/nsq-rpc/src/common"
	"github.com/PumpkinSeed/nsq-rpc/src/consumer"
	"github.com/PumpkinSeed/nsq-rpc/src/producer"
	"github.com/PumpkinSeed/nsq-rpc/src/utils"
	nsq "github.com/nsqio/go-nsq"
)

const (
	topic       = "lines"
	ch          = "metrics"
	chBack      = "back"
	log         = false
	storagePath = "test.db"
)

var consume bool
var amount int

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.BoolVar(&consume, "consume", false, "Consume the queue")
	flag.IntVar(&amount, "amount", 10, "Amount of the test messages")
	flag.Parse()
}

func main() {
	storage := common.NewStorage(storagePath)

	// read from recovery topic
	if consume {
		go func() {
			for i := 0; i < amount; i++ {
				produceHandler(utils.RandStringRunes(10), storage)
			}
		}()
		consumeHandler(storage)
	} else {
		for i := 0; i < amount; i++ {
			produceHandler(utils.RandStringRunes(10), storage)
		}
	}
}

func consumeHandler(s common.Storage) {
	config := consumer.NewConfig(topic, ch, false, log, &messageHandler{s})
	consumer.Start(config)

	// 2. Message Handler produce back to the lines+id topic
}

func produceHandler(msg string, s common.Storage) error {
	config := producer.NewConfig(topic, log)

	m, err := common.NewMessage(msg)
	if err != nil {
		return err
	}
	data, err := common.Encode(m)
	if err != nil {
		return err
	}
	producer.Write(data, config)

	s.Write(utils.GetTopicName(topic, m.ID))
	// recovery database for handle replies after a failure on the service

	configConsumer := consumer.NewConfig(utils.GetTopicName(topic, m.ID), chBack, true, log, &messageHandlerBack{s})
	go consumer.Start(configConsumer)
	// 3b. Start to wait for reply in lines+{id}

	return nil
}

type messageHandler struct {
	storage common.Storage
}

func (h *messageHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// returning an error results in the message being re-enqueued
		// a REQ is sent to nsqd
		return errors.New("body is blank re-enqueue message")
	}

	// Let's log our message!
	msg, err := common.Decode(m.Body)
	if err != nil {
		return err
	}

	singleTopic := utils.GetTopicName(topic, msg.ID)

	config := producer.NewConfig(singleTopic, log)

	rs, err := common.NewReadySignal(msg.ID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(rs)
	if err != nil {
		return err
	}

	producer.Write(data, config)

	h.storage.Remove(singleTopic)
	fmt.Println(msg)

	// Returning nil signals to the consumer that the message has
	// been handled with success. A FIN is sent to nsqd
	return nil
}

type messageHandlerBack struct {
	storage common.Storage
}

func (h *messageHandlerBack) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// returning an error results in the message being re-enqueued
		// a REQ is sent to nsqd
		return errors.New("body is blank re-enqueue message")
	}

	var rs common.ReadySignal
	err := json.Unmarshal(m.Body, &rs)
	if err != nil {
		return err
	}
	cmd := nsq.UnRegister(utils.GetTopicName(topic, rs.ID), "")
	cnf := nsq.NewConfig()
	addr := m.NSQDAddress

	conn := nsq.NewConn(addr, cnf, &common.Delegate{})

	err = conn.WriteCommand(cmd)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("Write: %s\n", addr)
	fmt.Println(utils.GetTopicName(topic, rs.ID))

	// Returning nil signals to the consumer that the message has
	// been handled with success. A FIN is sent to nsqd
	return nil
}
