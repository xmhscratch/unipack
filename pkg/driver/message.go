package driver

import (
	"bytes"
	"container/heap"
	"time"
	"unipack/pkg/fbgen-go/schema/websock"

	flatbuffers "github.com/google/flatbuffers/go"
)

func (ctx MessageItem) Index() int64 {
	return time.Now().UTC().Unix()
}

func (h MessageBufferStack) Len() int {
	return len(h)
}
func (h MessageBufferStack) Less(i int, j int) bool { return h[i].Index() < h[j].Index() }
func (h MessageBufferStack) Swap(i int, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MessageBufferStack) Push(i any) {
	*h = append(*h, i.(*MessageItem))
}

func (h *MessageBufferStack) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func NewMsgBufStack() *MessageBufferStack {
	var ctx *MessageBufferStack = &MessageBufferStack{}
	heap.Init(ctx)

	return ctx
}

// ===============================================

func NewMessage(namespace string, timestamp int64, data []uint8) []byte {
	size := len(data) + flatbuffers.SizeUint8
	builder := flatbuffers.NewBuilder(size)

	var offsetEnd flatbuffers.UOffsetT
	{
		websock.MessageStart(builder)
		websock.MessageAddEvent(builder, int16(SocketBinaryMessageEvent))
		websock.MessageAddTimestamp(builder, timestamp)
		websock.MessageAddNamespace(builder, builder.CreateString(namespace))
		websock.MessageAddMessage(builder, builder.CreateByteVector(data))
		offsetEnd = websock.MessageEnd(builder)
		websock.FinishMessageBuffer(builder, offsetEnd)
	}

	return builder.FinishedBytes()
}

func ReadMessage(src []byte, offset flatbuffers.UOffsetT, sizePrefix bool) *TMessage {
	buf := bytes.NewBuffer(src)
	var wsm *websock.Message
	if sizePrefix {
		wsm = websock.GetSizePrefixedRootAsMessage(buf.Bytes(), offset)
	} else {
		wsm = websock.GetRootAsMessage(buf.Bytes(), offset)
	}
	return &TMessage{
		Event:     TSocketMessageEvent(wsm.Event()),
		Namespace: string(wsm.Namespace()),
		Timestamp: wsm.Timestamp(),
		Message:   wsm.MessageBytes(),
	}
}
