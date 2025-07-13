package collector

import (
	"fmt"
	"strings"
	"time"
)

const TIMEOUT_READ_MESSAGE = time.Second * 30
const UUIDFormat = "%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x"
const UUIDNamespace = "6ba7b8109dad11d180b400c02fd430c8"

type TMessageBytes []uint8

func (u TMessageBytes) MarshalJSON() ([]byte, error) {
	var result string
	if u == nil {
		result = "null"
	} else {
		result = strings.Join(strings.Fields(fmt.Sprintf("%d", u)), ",")
	}
	return []byte(result), nil
}

type SocketConnection struct {
	UUID      string
	SessionId string
}

type SocketMessage struct {
	Event     int16   `json:"event"`
	Timestamp int64   `json:"timestamp"`
	Message   []uint8 `json:"message"`
}
