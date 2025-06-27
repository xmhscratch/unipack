package driver

import (
	"bytes"
	"encoding/json"
	"io"
)

type IDriverRW interface {
	io.ReadWriteCloser
	Send() error
}

type DriverStdout struct {
	IDriverRW
	buf     *bytes.Buffer
	Message json.RawMessage
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
}
