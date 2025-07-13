package collector

import (
	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/fiber/v2"
)

func HandleConnect(app *fiber.App) func(*socketio.EventPayload) {
	return func(ep *socketio.EventPayload) {}
}
