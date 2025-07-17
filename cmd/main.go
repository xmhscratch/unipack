package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
	"unipack/pkg/stat"
	"unipack/pkg/uni"

	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	// "github.com/gofiber/fiber/v2/middleware/filesystem"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("'run' commands not specified")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		runCmd := flag.NewFlagSet("run", flag.ExitOnError)

		go func() {
			MountTar(runCmd)
		}()
		<-StartServer()
	case "server":
		break
	default:
		os.Exit(1)
	}
}

func MountTar(runCmd *flag.FlagSet) os.Signal {
	var (
		mainFile   string
		mountPoint string
		entries    uni.StringMapFlag
	)
	runCmd.StringVar(&mainFile, "main-file", "main", "application main execution binary file")
	runCmd.StringVar(&mountPoint, "mount-point", "", "(optional) define default mount point on the host machine")
	runCmd.Var(&entries, "entries", "(optional) mounting external files or folders at runtime")
	runCmd.Parse(os.Args[2:])

	// fmt.Println(mainFile, mountPoint, entries)
	// fmt.Println(runCmd.Arg(0))

	if runCmd.Arg(0) == "" {
		panic("usage: Tar-FILE not specified")
	}

	tarFile, err := filepath.Abs(runCmd.Arg(0))
	if err != nil {
		panic(err)
	}

	return (&uni.VFSRoot{
		TarFile:    tarFile,
		MainFile:   mainFile,
		MountPoint: mountPoint,
	}).Serve()
}

func StartServer() chan os.Signal {
	app := fiber.New()

	app.Use(cors.New())

	// socketio.On(socketio.EventConnect, stat.HandleConnect(app))
	// socketio.On(socketio.EventDisconnect, stat.HandleDisconnect(app))
	socketio.On(socketio.EventClose, stat.HandleClose(app))
	// socketio.On(socketio.EventError, stat.HandleError(app))
	socketio.On(socketio.EventPing, stat.HandlePing(app))

	app.Use("/viewer", stat.NewStorageHandler("/home/web/repos/unipack/dist/viewer", stat.StorageConfig{
		Compress:      false,
		ByteRange:     true,
		Browse:        false,
		Download:      false,
		CacheDuration: 10 * time.Second,
		MaxAge:        3600,
	}))
	app.Use("/ws/*", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/:namespace", socketio.New(stat.NewSocketRoute(app)))
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("*", func(c *fiber.Ctx) error {
		return c.SendString("page not found!")
	})

	_, port, err := stat.ParseHostPort("")
	if err != nil {
		panic(err)
	}
	go app.Listen(net.JoinHostPort("0.0.0.0", port))

	return stat.WaitTermination()
}
