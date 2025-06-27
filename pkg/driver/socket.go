package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func (ctx *SocketStdout) Send() error {
	// buf, err := ctx.Message.MarshalJSON()
	fmt.Println(string(ctx.Message))
	// ctx.message = json.RawMessage(ctx.buf.Bytes())
	return nil
}

func (ctx *SocketStdout) Write(p []byte) (b int, err error) {
	if ctx.buf == nil {
		ctx.buf = bytes.NewBuffer(make([]byte, 1024))
	}
	return ctx.buf.Write(p)
}

func (ctx *SocketStdout) Close() (err error) {
	ctx.Message = json.RawMessage(ctx.buf.Bytes())
	return ctx.Send()
}
