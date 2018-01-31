package rpc

import (
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
	var testBody = "o"

	var e = Envelope{
		Method:        "put",
		ReplyTo:       "rsc",
		CorrelationID: 322232,
		ExpiresAt:     int64(time.Now().Nanosecond()),
		Body:          []byte(testBody),
	}

	encoded := e.Encode()

	e2, err := Decode(encoded)
	if err != nil {
		t.Error(err)
	}

	if string(e2.Body) != testBody {
		t.Errorf("Body should be %s, instead of %s", string(e2.Body), testBody)
	}
}

func TestExpired(t *testing.T) {
	var testBody = "o"

	var e = Envelope{
		Method:        "put",
		ReplyTo:       "rsc",
		CorrelationID: 322232,
		ExpiresAt:     int64(time.Now().Nanosecond()),
		Body:          []byte(testBody),
	}

	if !e.Expired() {
		t.Error("Expired should be true")
	}
}

func TestReply(t *testing.T) {
	var req = Envelope{
		Method:        "put",
		ReplyTo:       "rsc",
		CorrelationID: 322232,
		ExpiresAt:     int64(time.Now().Nanosecond()),
		Body:          []byte("request"),
	}

	resp := req.Reply([]byte("response"), nil)

	if string(resp.Body) != "response" {
		t.Errorf("Body should be %s, instead of %s", string(resp.Body), "response")
	}
}
