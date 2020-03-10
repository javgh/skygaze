package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/javgh/skygaze/skygazer"
)

func installSignalHandlers(cancel context.CancelFunc) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	<-c
	cancel()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go installSignalHandlers(cancel)

	skygazer := skygazer.New()
	err := skygazer.Listen(ctx, "skygaze.sock")
	if err != nil {
		log.Fatal(err)
	}
}
