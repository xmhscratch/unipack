package collector

import (
	"fmt"

	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/fiber/v2"
)

func HandleError(app *fiber.App) func(*socketio.EventPayload) {
	return func(ep *socketio.EventPayload) {
		fmt.Println(ep.Error)
	}
}
