package stat

import (
	"fmt"

	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/groupcache/lru"
)

func HandleClose(app *fiber.App) func(*socketio.EventPayload) {
	return func(ep *socketio.EventPayload) {
		connPool := socketPool.Get().(*lru.Cache)
		connPool.Remove(ep.Kws.GetAttribute("namespace"))
		ep.Kws.Conn.Close()
		fmt.Println(ep.Error)
	}
}
