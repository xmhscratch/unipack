package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"runtime"
	"strconv"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"golang.org/x/sync/semaphore"
)

func NewSocketStdout() (*SocketStdout, error) {
	ctx := &SocketStdout{
		dbOpts:    badger.DefaultOptions(STD_OUTPUT_PATH),
		idleSince: time.Now(),
		DriverStdout: DriverStdout{
			Messages: NewMsgBufStack(),
		},
	}

	db, err := ctx.OpenDb()
	if err != nil {
		return nil, err
	}
	go ctx.background(db)
	ctx.db = db

	return ctx, err
}

func (ctx *SocketStdout) background(db *badger.DB) {
	gc := time.After(5 * time.Second)
	flush := time.After(1 * time.Second)
	idle := make(chan struct{})

	tck := time.NewTicker(5 * time.Millisecond)
	defer tck.Stop()

loopEnd:
	for range tck.C {
	loopJob:
		go func() {
			if time.Since(ctx.idleSince).Milliseconds() >= (10 * time.Second).Milliseconds() {
				idle <- struct{}{}
			}
		}()

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
					item := ctx.Messages.Pop().(*MsgItem)
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

				goto loopJob
			}
			// println("flush")
			// goto loopJob

		case <-gc:
			err := db.RunValueLogGC(0.7)
			if err == nil {
				gc = time.After(5 * time.Second)
				// println("gc")
				goto loopJob
			}

		case <-idle:
			db.Close()
			// println("idle")
			break loopEnd
		}
	}
}

func (ctx *SocketStdout) OpenDb() (*badger.DB, error) {
	db, err := badger.Open(ctx.dbOpts)
	if err != nil {
		return nil, err
	}
	return db, err
}

func (ctx *SocketStdout) Write(p []byte) (b int, err error) {
	if ctx.buf == nil {
		ctx.buf = bytes.NewBuffer(make([]byte, 0))
	}
	return ctx.buf.Write(p)
}

func (ctx *SocketStdout) Close() (err error) {
	if ctx.db.IsClosed() {
		db, err := ctx.OpenDb()
		if err != nil {
			return err
		}
		ctx.db = db
	}
	ctx.Messages.Push(&MsgItem{Value: json.RawMessage(ctx.buf.Bytes())})
	return err
}
