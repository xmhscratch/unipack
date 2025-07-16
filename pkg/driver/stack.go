package driver

import (
	"container/heap"
	"time"
)

func (ctx MsgItem) Index() int64 {
	return time.Now().UTC().Unix()
}

func (h MsgBufStack) Len() int {
	return len(h)
}
func (h MsgBufStack) Less(i int, j int) bool { return h[i].Index() < h[j].Index() }
func (h MsgBufStack) Swap(i int, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MsgBufStack) Push(i any) {
	*h = append(*h, i.(*MsgItem))
}

func (h *MsgBufStack) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func NewMsgBufStack() *MsgBufStack {
	var ctx *MsgBufStack = &MsgBufStack{}
	heap.Init(ctx)

	return ctx
}
