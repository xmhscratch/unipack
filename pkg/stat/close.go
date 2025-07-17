package stat

import (
	"fmt"

	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/fiber/v2"
)

func HandleClose(app *fiber.App) func(*socketio.EventPayload) {
	return func(ep *socketio.EventPayload) {
		socketManager.Remove(ep.Kws.UUID)
		ep.Kws.Conn.Close()
		fmt.Println(ep.Error)
	}
}
