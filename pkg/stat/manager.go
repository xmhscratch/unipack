package stat

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"unipack/pkg/driver"

	"github.com/golang/groupcache/lru"
	flatbuffers "github.com/google/flatbuffers/go"
)

var connPool = lru.New(10)
var socketPool *sync.Pool = NewSocketPool()

func NewSocketPool() *sync.Pool {
	connPool.OnEvicted = func(key lru.Key, value interface{}) {
		sock := (value.(*sync.Pool)).Get().(*SocketConnection)
		(value.(*sync.Pool)).Put(sock)
	}

	return &sync.Pool{
		New: func() interface{} {
			return connPool
		},
	}
}

func NewDataStreamingSocket() {
	if _, err := os.Stat(UNIX_SOCKET_PATH); err == nil {
		os.Remove(UNIX_SOCKET_PATH)
	}

	listener, err := net.Listen("unix", UNIX_SOCKET_PATH)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("listening on", UNIX_SOCKET_PATH)

	for {
		conn, err := listener.Accept()
		if err != nil {
			// log.Println(err)
			continue
		}

		respBytes, err := ReadSocketRawData(conn)
		if err != nil {
			continue
		}

		msg := driver.ReadMessage(respBytes, flatbuffers.UOffsetT(0), false)
		println(msg)
	}
}

func ReadSocketRawData(rd io.Reader) ([]uint8, error) {
	var (
		err error
		buf *bytes.Buffer = bytes.NewBuffer(make([]byte, 0))
	)

	reader := bufio.NewReader(rd)

	for {
		c, err := reader.ReadByte()
		if err != nil && err == io.EOF {
			break
		}
		err = buf.WriteByte(c)
		if err != nil {
			return buf.Bytes(), fmt.Errorf("error getting input: %s", err)
		}
	}
	return buf.Bytes(), err
}
