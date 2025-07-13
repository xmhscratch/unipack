package main

import (
	"fmt"
	"testing"
	"time"
)

func TestMessageReceive(t *testing.T) {
	// var msg string = ""
	var msgChan chan string = make(chan string)

	ticker := time.NewTicker(time.Duration(5000) * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				msgChan <- "asdasd"
			}
		}
	}()

	// stopMsg:
	for range time.Tick(time.Duration(10) * time.Millisecond) {
		fmt.Println("check...")
	waitMsg:
		select {
		case msg := <-msgChan:
			fmt.Printf("message received: %+v\n", msg)
		case <-time.After(time.Duration(500) * time.Millisecond):
			fmt.Println("timeout")
			break waitMsg
		}
	}
}
