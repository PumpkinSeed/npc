package rpc

import (
	"bytes"
	"encoding/json"
	"time"
)

var (
	headerSeparator = []byte{10} //new line
)

// Envelope arround message for request response communication over nsq.
type Envelope struct {
	// method to call on the server side
	Method string `json:"m,omitempty"`
	// nsq topic to send reply to
	ReplyTo string `json:"r,omitempty"`
	// connection between request and response
	CorrelationID uint32 `json:"c,omitempty"`
	// unix timestamp when message expires, after that should be dropped
	ExpiresAt int64 `json:"x,omitempty"`
	// applicationn error reponse, if server side failed and Body is missing
	Error string `json:"e,omitempty"`
	// message body
	Body []byte `json:"-"`
}

// Reply creates reply Envelope from request Envelope.
func (m *Envelope) Reply(body []byte, err error) *Envelope {
	// refresh Envelope with the new body
	e := &Envelope{
		CorrelationID: m.CorrelationID,
		Body:          body,
	}

	// attach err if it is not nil
	if err != nil {
		e.Error = err.Error()
	}
	return e
}

// Expired returns true if message expired.
func (m *Envelope) Expired() bool {
	// checks ExpiresAt is nil of int
	if m.ExpiresAt <= 0 {
		return false
	}

	// checks it is expired or not
	return time.Now().Unix() > m.ExpiresAt
}

// Decode decodes envelope from bytes.
func Decode(buf []byte) (*Envelope, error) {
	// get the body chunk
	parts := bytes.SplitN(buf, headerSeparator, 2)

	// parse the json to the envelope
	e := &Envelope{}
	if err := json.Unmarshal(parts[0], e); err != nil {
		return nil, err
	}

	// put the body into the envelope
	if len(parts) > 1 {
		e.Body = parts[1]
	}
	return e, nil
}

// Encode encodes envelope into bytes for putting on wire.
func (m *Envelope) Encode() []byte {
	// encode the Envelope to JSON without the body
	buf, _ := json.Marshal(m)

	// add new line separator
	buf = append(buf, headerSeparator...)

	// add body to the message
	buf = append(buf, m.Body...)
	return buf
}
