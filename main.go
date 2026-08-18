package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"serverSideEvents/app"
)

func main() {

	app := app.New()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := app.Start(ctx)

	if err != nil {
		fmt.Printf("Error starting application: %v\n", err)
	}
}
