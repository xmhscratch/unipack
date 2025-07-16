package collector

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func NewSocketRoute(app *fiber.App) func(*socketio.Websocket) {
	return func(kws *socketio.Websocket) {
		var (
			err       error
			namespace string = kws.Params("namespace")
		)

		kws.SetAttribute("namespace", namespace)

		soc := NewSocketConnection(namespace)
		kws.UUID = soc.Get().(*SocketConnection).UUID

	checkNewMessage:
		for range time.Tick(time.Duration(10) * time.Millisecond) {
			var msgChan chan *TSocketMessage = make(chan *TSocketMessage)
			defer close(msgChan)

			go func() {
				var (
					tp int
					pl []byte
				)

				if tp, pl, err = kws.Conn.ReadMessage(); err != nil {
					// kws.Fire(socketio.EventError, []byte(err.Error()))
					fmt.Println(err.Error())
					msgChan <- nil
					return
				}

				if tp != 1 {
					msgChan <- nil
					return
				}

				{
					var msg *TSocketMessage
					if err = json.Unmarshal(pl, &msg); err != nil {
						// kws.Fire(socketio.EventError, []byte{})
						fmt.Println(err.Error())
						msgChan <- nil
						return
					}
					msgChan <- msg
				}
			}()

		readMsg:
			select {
			case msg := <-msgChan:
				if msg == nil {
					msg = &TSocketMessage{
						Event: websocket.CloseMessage,
					}
				}

				switch msg.Event {
				case websocket.PingMessage:
					kws.Fire(socketio.EventPing, []byte{})
				case websocket.TextMessage:
					kws.Fire(socketio.EventMessage, []byte{})
				case websocket.CloseMessage:
					kws.Fire(socketio.EventClose, []byte{})
					break checkNewMessage
				// kws.Fire(socketio.EventPing, []byte{})
				// case websocket.BinaryMessage:
				// 	if err := kws.Conn.WriteJSON(string(p)); err != nil {
				// 		kws.Fire(socketio.EventError, []byte(err.Error()))
				// 	}
				default:
					kws.Fire(socketio.EventClose, []byte{})
					break checkNewMessage
				}

				break readMsg
			case <-time.After(TIMEOUT_READ_MESSAGE):
				fmt.Println("read message timeout")
				break readMsg
			}
		}
	}
}
