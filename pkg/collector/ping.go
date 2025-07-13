package collector

import (
	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/fiber/v2"
)

func HandlePing(app *fiber.App) func(*socketio.EventPayload) {
	return func(ep *socketio.EventPayload) {
		// var (
		// 	err  error
		// 	sock *SocketConnection
		// )

		// if s, ok := socketManager.Get(ep.Kws.UUID); !ok {
		// 	ep.Kws.Fire(socketio.EventClose, []byte{})
		// 	return
		// } else {
		// 	sock = s.(*sync.Pool).Get().(*SocketConnection)
		// }

		// go func() {
		// stopMessage:
		// 	for range time.Tick(time.Duration(10) * time.Millisecond) {
		// if sock.Subscriber == nil {
		// 	break stopMessage
		// }

		// if !ep.Kws.IsAlive() {
		// 	break stopMessage
		// }

		// var (
		// 	msgChan <-chan *redis.Message = sock.Subscriber.Channel()
		// 	msg     *redis.Message        = <-msgChan
		// )

		// // fmt.Println(msg)

		// if msg == nil {
		// 	break stopMessage
		// }

		// if err := ep.Kws.Conn.WriteJSON(&filesrv.SocketMessage{
		// 	Event:   websocket.TextMessage,
		// 	Payload: msg.Payload,
		// }); err != nil {
		// 	fmt.Println(err)
		// 	break stopMessage
		// 	// ep.Kws.Fire(socketio.EventError, []byte(err.Error()))
		// }
		// fmt.Println(ep.Kws.UUID, ep.Kws.IsAlive())
		// }
		// }()
	}
}
