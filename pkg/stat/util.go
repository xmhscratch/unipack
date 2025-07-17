package stat

import (
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func ParseHostPort(address string) (string, string, error) {
	if !strings.Contains(address, ":") {
		return address, "3113", nil
	}
	host, port, err := net.SplitHostPort(address)
	return host, port, err
}

func WaitTermination() chan os.Signal {
	var c chan os.Signal = make(chan os.Signal, 4)

	signal.Notify(
		c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				return
			}
		}
	}()

	return c
}
