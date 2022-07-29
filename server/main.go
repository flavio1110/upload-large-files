package main

import (
	"context"
	"files-api/api"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	port := "8080"

	api := api.NewApiServer(port)

	go func() {
		fmt.Printf("listening on port %q. waiting for shutdown...\n", port)
		api.Start()
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT)
	<-exit

	fmt.Println("Shutting down...")
	if err := api.Stop(context.Background()); err != nil {
		fmt.Printf("fail to gacefully shutdown the http server! %s\n", err)
	}

	fmt.Println("Bye!")
}
