package rpc

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
	var testBody = "msg"

	var e = Envelope{
		Method:        "put",
		ReplyTo:       "rsc",
		CorrelationID: 322232,
		ExpiresAt:     int64(time.Now().Nanosecond()),
		Body:          []byte(testBody),
	}

	encoded := e.Encode()

	body := strings.Split(string(encoded), string(headerSeparator))[1]

	if body != testBody {
		t.Errorf("Body should be %s, instead of %s", body, testBody)
	}
}

func TestDecode(t *testing.T) {
	var e = Envelope{
		Method:        "put",
		ReplyTo:       "rsc",
		CorrelationID: 322232,
		ExpiresAt:     int64(time.Now().Nanosecond()),
		Body:          []byte("o"),
	}

	encoded := e.Encode()

	e2, err := Decode(encoded)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(e2.Body)
}
