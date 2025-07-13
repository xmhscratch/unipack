package main

import (
	"encoding/json"
	"fmt"
	"time"
	"unipack/pkg/collector"

	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func NewSocketRoute(app *fiber.App) func(*socketio.Websocket) {
	return func(kws *socketio.Websocket) {
		var (
			err       error
			sessionId string = kws.Params("sessionId")
		)

		kws.SetAttribute("session_id", sessionId)

		soc := collector.NewSocketConnection(sessionId)
		kws.UUID = soc.Get().(*collector.SocketConnection).UUID

	checkNewMessage:
		for range time.Tick(time.Duration(10) * time.Millisecond) {
			var msgChan chan *collector.SocketMessage = make(chan *collector.SocketMessage)
			defer close(msgChan)

			go func() {
				var (
					tp int
					pl []byte
				)

				if tp, pl, err = kws.Conn.ReadMessage(); err != nil {
					// kws.Fire(socketio.EventError, []byte(err.Error()))
					fmt.Println(err.Error())
					// msgChan <- nil
					return
				}

				if tp != 1 {
					// msgChan <- nil
					return
				}

				{
					var msg *collector.SocketMessage
					if err = json.Unmarshal(pl, &msg); err != nil {
						// kws.Fire(socketio.EventError, []byte{})
						fmt.Println(err.Error())
						// msgChan <- nil
						return
					}
					msgChan <- msg
				}
			}()

		readMsg:
			select {
			case msg := <-msgChan:
				if msg == nil {
					msg = &collector.SocketMessage{
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
			case <-time.After(collector.TIMEOUT_READ_MESSAGE):
				fmt.Println("read message timeout")
				break readMsg
			}
		}
	}
}
