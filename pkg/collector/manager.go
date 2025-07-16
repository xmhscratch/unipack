package collector

import (
	"sync"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/golang/groupcache/lru"
)

var socketManager *lru.Cache

func NewSocketConnection(namespace string) *sync.Pool {
	var sockUUID string = GenerateV5(namespace, UUIDNamespace, UUIDNamespace)
	ctx := &sync.Pool{
		New: func() interface{} {
			var (
				err  error
				db   *badger.DB
				opts badger.Options = badger.DefaultOptions(STD_OUTPUT_PATH)
			)

			db, err = badger.Open(opts)
			if err != nil {
				return err
			}

			return &SocketConnection{
				UUID:      sockUUID,
				Namespace: namespace,
				Db:        db,
			}
		},
	}

	if socketManager == nil {
		socketManager = lru.New(10)
		socketManager.OnEvicted = func(key lru.Key, value interface{}) {
			sock := (value.(*sync.Pool)).Get().(*SocketConnection)
			defer sock.Db.Close()
			(value.(*sync.Pool)).Put(sock)
		}
	}

	socketManager.Add(sockUUID, ctx)
	return ctx
}
