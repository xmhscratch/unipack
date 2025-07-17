package stat

import (
	"io/fs"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/gofiber/fiber/v2"
)

const STD_OUTPUT_PATH = "/var/log/unipack"
const TIMEOUT_READ_MESSAGE = time.Second * 30
const UUIDFormat = "%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x"
const UUIDNamespace = "6ba7b8109dad11d180b400c02fd430c8"

type SocketConnection struct {
	UUID      string
	Namespace string
	Db        *badger.DB
}

type TSocketEvent (int16)

// Source @url:https://github.com/gorilla/websocket/blob/master/conn.go#L61
// The message types are defined in RFC 6455, section 11.8.
const (
	_                        TSocketEvent = iota
	SocketTextMessageEvent   TSocketEvent = 1
	SocketBinaryMessageEvent TSocketEvent = 2
	SocketCloseMessageEvent  TSocketEvent = 8
	SocketPingMessageEvent   TSocketEvent = 9
	SocketPongMessageEvent   TSocketEvent = 10
)

type TSocketMessage struct {
	Event     TSocketEvent `json:"event"`
	Namespace string       `json:"namespace"`
	Timestamp int64        `json:"timestamp"`
	Message   []uint8      `json:"message"`
}

// Config defines the config for middleware.
type StorageConfig struct {
	// FS is the file system to serve the static files from.
	// You can use interfaces compatible with fs.FS like embed.FS, os.DirFS etc.
	//
	// Optional. Default: nil
	FS fs.FS

	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(c *fiber.Ctx) bool

	// ModifyResponse defines a function that allows you to alter the response.
	//
	// Optional. Default: nil
	ModifyResponse fiber.Handler

	// NotFoundHandler defines a function to handle when the path is not found.
	//
	// Optional. Default: nil
	NotFoundHandler fiber.Handler

	// The names of the index files for serving a directory.
	//
	// Optional. Default: []string{"index.html"}.
	IndexNames []string `json:"index"`

	// Expiration duration for inactive file handlers.
	// Use a negative time.Duration to disable it.
	//
	// Optional. Default: 10 * time.Second.
	CacheDuration time.Duration `json:"cache_duration"`

	// The value for the Cache-Control HTTP-header
	// that is set on the file response. MaxAge is defined in seconds.
	//
	// Optional. Default: 0.
	MaxAge int `json:"max_age"`

	// When set to true, the server tries minimizing CPU usage by caching compressed files.
	// This works differently than the github.com/gofiber/compression middleware.
	//
	// Optional. Default: false
	Compress bool `json:"compress"`

	// When set to true, enables byte range requests.
	//
	// Optional. Default: false
	ByteRange bool `json:"byte_range"`

	// When set to true, enables directory browsing.
	//
	// Optional. Default: false.
	Browse bool `json:"browse"`

	// When set to true, enables direct download.
	//
	// Optional. Default: false.
	Download bool `json:"download"`
}

// ConfigDefault is the default config
var StorageConfigDefault = StorageConfig{
	IndexNames:    []string{"index.html"},
	CacheDuration: 10 * time.Second,
}
