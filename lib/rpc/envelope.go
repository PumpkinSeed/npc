package rpc

import (
	"bytes"
	"encoding/binary"
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
	e := &Envelope{
		CorrelationID: m.CorrelationID,
		Body:          body,
	}
	if err != nil {
		e.Error = err.Error()
	}
	return e
}

// Expired returns true if message expired.
func (m *Envelope) Expired() bool {
	if m.ExpiresAt <= 0 {
		return false
	}
	return time.Now().Unix() > m.ExpiresAt
}

// Decode decodes envelope from bytes.
func Decode(buf []byte) (*Envelope, error) {
	parts := bytes.SplitN(buf, headerSeparator, 2)
	e := &Envelope{}
	if err := json.Unmarshal(parts[0], e); err != nil {
		return nil, err
	}
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

	// add body
	buf = append(buf, m.Body...)
	return buf
}

// encode implements binary encoding. Currently inactive.
// 0 - encodes request (Method, CorrelationID, ReplyTo, ExpiresAt, Body)
// 1 - encodes reponse (CorrelationID, Error, Body)
func (m *Envelope) encode(et int) []byte {
	lenBody := len(m.Body)
	switch et {
	case 0:
		lenType := len(m.Method)
		lenReplyTo := len(m.ReplyTo)
		res := make([]byte, 15+lenType+lenReplyTo+lenBody)
		//res[0] = uint8(0)
		res[1] = uint8(lenType)
		res[2] = uint8(lenReplyTo)
		binary.BigEndian.PutUint32(res[3:7], m.CorrelationID)
		binary.BigEndian.PutUint64(res[7:15], uint64(m.ExpiresAt))
		pos := 15
		copy(res[pos:pos+lenType], []byte(m.Method))
		pos = pos + lenType
		copy(res[pos:pos+lenReplyTo], []byte(m.ReplyTo))
		pos = pos + lenReplyTo
		copy(res[pos:pos+lenBody], m.Body)
		return res
	case 1:
		lenError := len(m.Error)
		res := make([]byte, 6+lenError+lenBody)
		res[0] = uint8(1)
		res[1] = uint8(lenError)
		binary.BigEndian.PutUint32(res[2:6], m.CorrelationID)
		pos := 6
		copy(res[pos:pos+lenError], []byte(m.Error))
		pos += lenError
		copy(res[pos:pos+lenBody], m.Body)
		return res
	}
	return nil
}

// decode implements binary decoding. Currently inactive.
func decode(buf []byte) (*Envelope, error) {
	m := &Envelope{}
	switch int(buf[0]) {
	case 0:
		lenType := int(buf[1])
		lenReplayTo := int(buf[2])
		m.CorrelationID = binary.BigEndian.Uint32(buf[3:7])
		m.ExpiresAt = int64(binary.BigEndian.Uint64(buf[7:15]))

		pos := 15
		m.Method = string(buf[pos : pos+lenType])
		pos += lenType
		m.ReplyTo = string(buf[pos : pos+lenReplayTo])
		pos += lenReplayTo
		m.Body = buf[pos:]

		return m, nil
	case 1:
		lenError := int(buf[1])
		m.CorrelationID = binary.BigEndian.Uint32(buf[2:6])
		pos := 6
		if lenError > 0 {
			m.Error = string(buf[pos : pos+lenError])
			pos += lenError
		}
		m.Body = buf[pos:]
		return m, nil
	}
	return nil, nil
}
