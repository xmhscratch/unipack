package driver

import (
	"bytes"
	"encoding/json"
	"io"
	"time"
)

const UNIX_SOCKET_PATH = "/run/unipack.sock"
const STD_OUTPUT_PATH = "/var/log/unipack"

type IDriverRW interface {
	io.ReadWriteCloser
	Send() error
}

type DriverStdout struct {
	IDriverRW
	buf      *bytes.Buffer
	Messages *MessageBufferStack
}

type EDriverErrCode (int)
type EDriverErrMessage json.RawMessage

const (
	_                EDriverErrCode = iota
	ERR_UNKNOWN      EDriverErrCode = 1
	ERR_CONN_TIMEOUT EDriverErrCode = 2
)

type DriverErrorMessage struct {
	EDriverErrCode
	EDriverErrMessage
}

type DriverStderr struct {
	IDriverRW
	Message DriverErrorMessage
}

type SocketStdout struct {
	DriverStdout
	idleSince time.Time
}

type TSocketMessageEvent (int16)

// Source @url:https://github.com/gorilla/websocket/blob/master/conn.go#L61
// The message types are defined in RFC 6455, section 11.8.
const (
	_                        TSocketMessageEvent = iota
	SocketTextMessageEvent   TSocketMessageEvent = 1
	SocketBinaryMessageEvent TSocketMessageEvent = 2
	SocketCloseMessageEvent  TSocketMessageEvent = 8
	SocketPingMessageEvent   TSocketMessageEvent = 9
	SocketPongMessageEvent   TSocketMessageEvent = 10
)

type TMessage struct {
	Event     TSocketMessageEvent
	Namespace string
	Timestamp int64
	Message   []uint8
}

type IMessageItem interface {
	Index() int64
}

type MessageItem struct {
	IMessageItem
	Value json.RawMessage
}

type MessageBufferStack ([]*MessageItem)
