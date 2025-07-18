package stat

import (
	"time"
	"unipack/pkg/driver"

	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/groupcache/lru"
	flatbuffers "github.com/google/flatbuffers/go"
)

func NewSocketRoute(app *fiber.App) func(*socketio.Websocket) {
	return func(kws *socketio.Websocket) {
		var (
			err       error
			namespace string = kws.Params("namespace")
			conn      *SocketConnection
		)

		kws.SetAttribute("namespace", namespace)

		connPool := socketPool.Get().(*lru.Cache)
		if _, ok := connPool.Get(namespace); !ok {
			conn = &SocketConnection{Namespace: namespace}
			connPool.Add(namespace, conn)

			socketPool.Put(connPool)
		}

		checkInterval := time.NewTicker(5 * time.Millisecond)
		defer checkInterval.Stop()

	checkNewMessage:
		for range checkInterval.C {
			var msg chan *driver.TMessage = make(chan *driver.TMessage)
			defer close(msg)

			go func() {
				var (
					tp int
					pl []byte
				)

				if tp, pl, err = kws.Conn.ReadMessage(); err != nil {
					kws.Fire(socketio.EventError, []byte(err.Error()))
					return
				}

				if tp < 1 || tp > 2 {
					return
				}
				msg <- driver.ReadMessage(pl, flatbuffers.UOffsetT(0), false)
			}()

			switch (<-msg).Event {
			case websocket.PingMessage:
				println("ping")
				kws.Fire(socketio.EventPing, []byte{})
			case websocket.TextMessage:
				println("text")
				kws.Fire(socketio.EventMessage, []byte{})
			case websocket.BinaryMessage:
				println("binary")
				kws.Fire(socketio.EventMessage, []byte{})
				// if err := kws.Conn.WriteJSON(string(p)); err != nil {
				// 	kws.Fire(socketio.EventError, []byte(err.Error()))
				// }
			case websocket.CloseMessage:
				kws.Fire(socketio.EventClose, []byte{})
				break checkNewMessage
			// kws.Fire(socketio.EventPing, []byte{})
			default:
				kws.Fire(socketio.EventClose, []byte{})
				break checkNewMessage
			}
		}
	}
}
