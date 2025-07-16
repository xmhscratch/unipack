package driver

import (
	"testing"
	"time"
)

func TestMechanic(t *testing.T) {
	gc := time.After(5 * time.Second)
	flush := time.After(1 * time.Second)
	idle := make(chan struct{})

	idleSince := time.Now()

	tck := time.NewTicker(5 * time.Millisecond)
	defer tck.Stop()

loopEnd:
	for range tck.C {
	loopJob:
		go func() {
			if time.Since(idleSince).Milliseconds() >= (10 * time.Second).Milliseconds() {
				idle <- struct{}{}
			}
		}()
		select {
		case <-flush:
			flush = time.After(1 * time.Second)
			println("flush")
			idleSince = time.Now()
			goto loopJob

		case <-gc:
			println("gc")
			gc = time.After(5 * time.Second)
			goto loopJob

		case <-idle:
			println("idle")
			break loopEnd
		}
	}
}
