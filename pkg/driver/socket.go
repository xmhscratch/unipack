package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"golang.org/x/sync/semaphore"
)

var db, _ = badger.Open(
	badger.DefaultOptions(STD_OUTPUT_PATH).
		WithSyncWrites(false).
		WithLoggingLevel(badger.WARNING),
)
var loggerPool *sync.Pool = NewLoggerPool()

func NewLoggerPool() *sync.Pool {
	return &sync.Pool{
		New: func() interface{} {
			if db.IsClosed() {
				var err error
				if db, err = badger.Open(
					badger.DefaultOptions(STD_OUTPUT_PATH).
						WithSyncWrites(false).
						WithLoggingLevel(badger.WARNING),
				); err != nil {
					panic(err)
				}
			}
			return db
		},
	}
}

func NewSocketStdout() (*SocketStdout, error) {
	ctx := &SocketStdout{
		idleSince: time.Now(),
		DriverStdout: DriverStdout{
			Messages: NewMsgBufStack(),
		},
	}
	go ctx.Background()

	return ctx, nil
}

func (ctx *SocketStdout) Background() {
	db := loggerPool.Get().(*badger.DB)
	defer loggerPool.Put(db)

	gc := time.After(5 * time.Second)
	flush := time.After(1 * time.Second)
	pushSocket := time.After(50 * time.Millisecond)
	idle := make(chan struct{})

	tck := time.NewTicker(5 * time.Millisecond)
	defer tck.Stop()

	go func() {
		for range tck.C {
			if time.Since(ctx.idleSince).Milliseconds() >= (10 * time.Second).Milliseconds() {
				idle <- struct{}{}
			}
		}
	}()

loopEnd:
	for range tck.C {
	loopJob:
		select {
		case <-flush:
			if ctx.Messages.Len() == 0 {
				goto loopJob
			}

			maxWorkers := runtime.GOMAXPROCS(0)
			todoCtx := context.TODO()
			sem := semaphore.NewWeighted(int64(maxWorkers))
			var errCh chan error = make(chan error, 1)

			if err := sem.Acquire(todoCtx, 1); err != nil {
				log.Printf("Failed to acquire semaphore: %v", err)
				break
			}

			go func(errCh chan error) {
				defer sem.Release(1)

				tx := db.NewTransaction(true)
				for range ctx.Messages.Len() {
					item := ctx.Messages.Pop().(*MessageItem)
					if err := tx.Set(
						[]byte(strconv.FormatInt(item.Index(), 10)),
						[]byte(item.Value),
					); err == badger.ErrTxnTooBig {
						_ = tx.Commit()
						tx = db.NewTransaction(true)
						_ = tx.Set(
							[]byte(strconv.FormatInt(item.Index(), 10)),
							[]byte(item.Value),
						)
					}
				}
				errCh <- tx.Commit()
			}(errCh)

			if err := sem.Acquire(todoCtx, int64(maxWorkers)); err != nil {
				log.Printf("Failed to acquire semaphore: %v", err)
				break
			}

			if err := <-errCh; err == nil {
				flush = time.After(1 * time.Second)

				ctx.buf.Reset()
				ctx.idleSince = time.Now()
			}
			println("flush")
			goto loopJob

		case <-gc:
			err := db.RunValueLogGC(0.7)
			if err == nil {
				gc = time.After(5 * time.Second)
				println("gc")
				goto loopJob
			}

		case <-pushSocket:
			ctx.PushSocket(6 * time.Hour)
			println("pushSocket")

		case <-idle:
			db.Close()
			println("idle")
			break loopEnd
		}
	}
}

func (ctx *SocketStdout) Write(p []byte) (b int, err error) {
	if ctx.buf == nil {
		ctx.buf = bytes.NewBuffer(make([]byte, 0))
	}
	return ctx.buf.Write(p)
}

func (ctx *SocketStdout) Close() (err error) {
	db := loggerPool.Get().(*badger.DB)
	defer loggerPool.Put(db)

	item := &MessageItem{Value: json.RawMessage(ctx.buf.Bytes())}
	ctx.Messages.Push(item)
	return err
}

func (ctx *SocketStdout) PushSocket(recentDuration time.Duration) error {
	db := loggerPool.Get().(*badger.DB)
	defer loggerPool.Put(db)

	conn, err := net.Dial("unix", UNIX_SOCKET_PATH)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	return db.View(func(tx *badger.Txn) error {
		iterOpts := badger.DefaultIteratorOptions
		iterOpts.PrefetchValues = false

		iter := tx.NewIterator(iterOpts)
		defer iter.Close()

		for iter.Rewind(); iter.Valid(); iter.Next() {
			item := iter.Item()

			timestamp, err := strconv.ParseInt(string(item.Key()), 10, 64)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			if time.Since(time.Unix(timestamp, 0).UTC()) > recentDuration {
				return nil
			}

			if err = item.Value(func(v []byte) error {
				src := bytes.NewReader(v)
				_, err = io.CopyBuffer(conn, src, make([]byte, 0))
				return err
			}); err != nil {
				return err
			}
		}
		return nil
	})
}
