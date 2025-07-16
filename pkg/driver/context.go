package driver

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

const STD_OUTPUT_PATH = "/var/log/unipack"

type IDriverRW interface {
	io.ReadWriteCloser
	Send() error
}

type DriverStdout struct {
	IDriverRW
	buf      *bytes.Buffer
	Messages *MsgBufStack
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
	db        *badger.DB
	dbOpts    badger.Options
	idleSince time.Time
}

type IMsgItem interface {
	Index() int64
}

type MsgItem struct {
	IMsgItem
	Value json.RawMessage
}

type MsgBufStack ([]*MsgItem)
