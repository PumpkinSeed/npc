package common

import nsq "github.com/nsqio/go-nsq"

type Delegate struct {
}

// OnResponse is called when the connection
// receives a FrameTypeResponse from nsqd
func (d *Delegate) OnResponse(conn *nsq.Conn, str []byte) {

}

// OnError is called when the connection
// receives a FrameTypeError from nsqd
func (d *Delegate) OnError(conn *nsq.Conn, str []byte) {

}

// OnMessage is called when the connection
// receives a FrameTypeMessage from nsqd
func (d *Delegate) OnMessage(conn *nsq.Conn, msg *nsq.Message) {

}

// OnMessageFinished is called when the connection
// handles a FIN command from a message handler
func (d *Delegate) OnMessageFinished(conn *nsq.Conn, msg *nsq.Message) {

}

// OnMessageRequeued is called when the connection
// handles a REQ command from a message handler
func (d *Delegate) OnMessageRequeued(conn *nsq.Conn, msg *nsq.Message) {

}

// OnBackoff is called when the connection triggers a backoff state
func (d *Delegate) OnBackoff(conn *nsq.Conn) {

}

// OnContinue is called when the connection finishes a message without adjusting backoff state
func (d *Delegate) OnContinue(conn *nsq.Conn) {

}

// OnResume is called when the connection triggers a resume state
func (d *Delegate) OnResume(conn *nsq.Conn) {

}

// OnIOError is called when the connection experiences
// a low-level TCP transport error
func (d *Delegate) OnIOError(conn *nsq.Conn, err error) {

}

// OnHeartbeat is called when the connection
// receives a heartbeat from nsqd
func (d *Delegate) OnHeartbeat(conn *nsq.Conn) {

}

// OnClose is called when the connection
// closes, after all cleanup
func (d *Delegate) OnClose(conn *nsq.Conn) {

}
