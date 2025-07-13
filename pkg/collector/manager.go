package collector

import (
	"sync"

	"github.com/golang/groupcache/lru"
)

var socketManager *lru.Cache

func NewSocketConnection(sessionId string) *sync.Pool {
	var sockUUID string = GenerateV5(sessionId, "fileKey", UUIDNamespace)
	ctx := &sync.Pool{
		New: func() interface{} {
			return &SocketConnection{
				UUID:      sockUUID,
				SessionId: sessionId,
			}
		},
	}

	if socketManager == nil {
		socketManager = lru.New(10)
		socketManager.OnEvicted = func(key lru.Key, value interface{}) {
			// sockUUID := key.(string)
			sock := (value.(*sync.Pool)).Get().(*SocketConnection)
			// // fmt.Println(sockUUID, sock)
			// if err := sock.Subscriber.Unsubscribe(context.TODO(), fileKey); err != nil {
			// }
			// if err := sock.Subscriber.Close(); err != nil {
			// }
			(value.(*sync.Pool)).Put(sock)
		}
	}

	socketManager.Add(sockUUID, ctx)
	return ctx
}
