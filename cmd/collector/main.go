package main

import (
	"net"

	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	// "github.com/gofiber/fiber/v2/middleware/filesystem"

	"unipack/pkg/collector"
)

func main() {
	app := fiber.New()

	app.Use(cors.New())

	// socketio.On(socketio.EventConnect, collector.HandleConnect(app))
	// socketio.On(socketio.EventDisconnect, collector.HandleDisconnect(app))
	socketio.On(socketio.EventClose, collector.HandleClose(app))
	// socketio.On(socketio.EventError, collector.HandleError(app))
	socketio.On(socketio.EventPing, collector.HandlePing(app))

	app.Use("/ws/*", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/:sessionId", socketio.New(NewSocketRoute(app)))

	app.Get("*", func(c *fiber.Ctx) error {
		return c.SendString("File not found!")
	})

	_, port, err := collector.ParseHostPort("")
	if err != nil {
		panic(err)
	}
	go app.Listen(net.JoinHostPort("0.0.0.0", port))

	collector.WaitTermination()
}
