package collector

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	badger "github.com/dgraph-io/badger/v4"
	flatbuffers "github.com/google/flatbuffers/go"

	"unipack/pkg/fbgen-go/schema/websock"
)

func HandlePing(app *fiber.App) func(*socketio.EventPayload) {
	return func(ep *socketio.EventPayload) {
		var (
			err  error
			sock *SocketConnection
		)

		if s, ok := socketManager.Get(ep.Kws.UUID); !ok {
			ep.Kws.Fire(socketio.EventClose, []byte{})
			return
		} else {
			sock = s.(*sync.Pool).Get().(*SocketConnection)
		}

		go func() {
			var cursor int64

			if err = sock.Db.View(func(tx *badger.Txn) error {
				iterOpts := badger.DefaultIteratorOptions
				iterOpts.PrefetchValues = false

				iter := tx.NewIterator(iterOpts)
				defer iter.Close()

				for iter.Rewind(); iter.Valid(); iter.Next() {
					item := iter.Item()
					timestamp, err := strconv.ParseInt(string(item.Key()), 10, 64)
					if err != nil {
						fmt.Println(err)
						return nil
					}

					if time.Since(time.Unix(timestamp, 0)) > 24*time.Hour {
						return nil
					}

					if err = item.Value(func(v []byte) error {
						fmt.Printf("key=%d, value=%s\n", timestamp, string(v))
						cursor = timestamp
						return nil
					}); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				fmt.Println(err)
				ep.Kws.Fire(socketio.EventClose, []byte{})
				return
			}

		stopMessage:
			for range time.Tick(time.Duration(10) * time.Millisecond) {
				// if sock.Subscriber == nil {
				// 	break stopMessage
				// }

				if !ep.Kws.IsAlive() {
					break stopMessage
				}

				if err = sock.Db.View(func(tx *badger.Txn) error {
					iterOpts := badger.DefaultIteratorOptions
					iterOpts.PrefetchValues = false

					iter := tx.NewIterator(iterOpts)
					defer iter.Close()

					for iter.Rewind(); iter.Valid(); iter.Next() {
						item := iter.Item()
						timestamp, err := strconv.ParseInt(string(item.Key()), 10, 64)
						if err != nil {
							fmt.Println(err)
							return nil
						}

						if time.Since(time.Unix(cursor, 0)) <= 0 {
							return nil
						}

						if err = item.Value(func(v []byte) error {
							msgBytes := NewMessage(sock.Namespace, timestamp, v)
							// fmt.Printf("key=%d, value=%s\n", timestamp, string(v))
							if err := ep.Kws.Conn.WriteMessage(websocket.BinaryMessage, msgBytes); err != nil {
								fmt.Println(err)
								// ep.Kws.Fire(socketio.EventError, []byte(err.Error()))
							}
							return nil
						}); err != nil {
							return err
						}
					}
					return nil
				}); err != nil {
					fmt.Println(err)
					ep.Kws.Fire(socketio.EventClose, []byte{})

					break stopMessage
				}

				// fmt.Println(ep.Kws.UUID, ep.Kws.IsAlive())
			}
		}()
	}
}

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

// func ReadMessage(src []byte, offset flatbuffers.UOffsetT, sizePrefix bool) *TSocketMessage {
// 	buf := bytes.NewBuffer(src)
// 	var wsm *websock.Message
// 	if sizePrefix {
// 		wsm = websock.GetSizePrefixedRootAsMessage(buf.Bytes(), offset)
// 	} else {
// 		wsm = websock.GetRootAsMessage(buf.Bytes(), offset)
// 	}
// 	return &TSocketMessage{
// 		Event:     TSocketEvent(wsm.Event()),
// 		Namespace: string(wsm.Namespace()),
// 		Timestamp: wsm.Timestamp(),
// 		Message:   wsm.MessageBytes(),
// 	}
// }
