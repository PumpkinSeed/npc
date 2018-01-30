package common

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Message struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func NewMessage(m string) (Message, error) {
	id, err := pseudoUUID()
	if err != nil {
		return Message{}, err
	}
	return Message{
		ID:        id,
		Message:   m,
		Timestamp: timestamp(),
	}, nil
}

func Encode(m Message) ([]byte, error) {
	return json.Marshal(m)
}

func Decode(data []byte) (Message, error) {
	var m Message
	err := json.Unmarshal(data, &m)
	return m, err
}

func pseudoUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

func timestamp() string {
	t := time.Now().UTC().UnixNano()
	return strconv.Itoa(int(t))
}
