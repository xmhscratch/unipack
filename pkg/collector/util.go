package collector

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

func WaitTermination() {
	exit := make(chan struct{})
	SignalC := make(chan os.Signal, 4)

	signal.Notify(
		SignalC,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	go func() {
		for s := range SignalC {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				close(exit)
				return
			}
		}
	}()

	<-exit
	os.Exit(0)
}
