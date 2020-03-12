package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/javgh/skygaze/broadcaster"
	"github.com/javgh/skygaze/skygazer"
)

const (
	broadcasterAddress = ":8023"
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

	broadcaster := broadcaster.New()
	go broadcaster.Serve(ctx, broadcasterAddress)

	skygazer := skygazer.New(broadcaster)
	err := skygazer.Listen(ctx, "skygaze.sock")
	if err != nil {
		log.Fatal(err)
	}
}
