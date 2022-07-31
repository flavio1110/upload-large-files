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

	if err := os.Mkdir("temp", 0604); err != nil {
		fmt.Println("failed to create temp folder", err)
	}
	defer func() {
		if err := os.RemoveAll("temp"); err != nil {
			fmt.Println("failed to delete temp folder")
		}
	}()

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
