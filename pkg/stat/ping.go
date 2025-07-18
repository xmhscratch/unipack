package stat

import (
	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/fiber/v2"
)

func HandlePing(app *fiber.App) func(*socketio.EventPayload) {
	return func(ep *socketio.EventPayload) {
		// var (
		// 	// err       error
		// 	namespace string = ep.Kws.GetAttribute("namespace").(string)
		// 	// sock      *SocketConnection
		// )

		// connPool := socketPool.Get().(*lru.Cache)
		// if c, ok := connPool.Get(namespace); !ok {
		// 	ep.Kws.Fire(socketio.EventClose, []byte{})
		// 	return
		// } else {
		// 	sock = c.(*SocketConnection)
		// }
		// socketPool.Put(connPool)

		// db := loggerPool.Get().(*badger.DB)
		// go func() {

		// stopMessage:
		// 	for range time.Tick(time.Duration(10) * time.Millisecond) {
		// 		// if sock.Subscriber == nil {
		// 		// 	break stopMessage
		// 		// }

		// 		if !ep.Kws.IsAlive() {
		// 			break stopMessage
		// 		}

		// 		if err = db.View(func(tx *badger.Txn) error {
		// 			iterOpts := badger.DefaultIteratorOptions
		// 			iterOpts.PrefetchValues = false

		// 			iter := tx.NewIterator(iterOpts)
		// 			defer iter.Close()

		// 			for rs {
		// 				timestamp, err := strconv.ParseInt(string(item.Key()), 10, 64)
		// 				if err != nil {
		// 					fmt.Println(err)
		// 					return nil
		// 				}

		// 				if time.Since(time.Unix(cursor, 0)) <= 0 {
		// 					return nil
		// 				}

		// 				if err = item.Value(func(v []byte) error {
		// 					msgBytes := NewMessage(sock.Namespace, timestamp, v)
		// 					// fmt.Printf("key=%d, value=%s\n", timestamp, string(v))
		// 					if err := ep.Kws.Conn.WriteMessage(websocket.BinaryMessage, msgBytes); err != nil {
		// 						fmt.Println(err)
		// 						// ep.Kws.Fire(socketio.EventError, []byte(err.Error()))
		// 					}
		// 					return nil
		// 				}); err != nil {
		// 					return err
		// 				}
		// 			}
		// 			return nil
		// 		}); err != nil {
		// 			fmt.Println(err)
		// 			ep.Kws.Fire(socketio.EventClose, []byte{})

		// 			break stopMessage
		// 		}

		// 		// fmt.Println(ep.Kws.UUID, ep.Kws.IsAlive())
		// 	}
		// }()
	}
}
