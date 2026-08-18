package app

import (
	"context"
	"fmt"
	"net/http"
	redisclient "serverSideEvents/client/redis"
	"serverSideEvents/handler"
	"time"
)

type App struct {
	router       http.Handler
	HandlerUtils *handler.Utils
	RedisClient  *redisclient.Client
}

func New() *App {

	handlerUtils := handler.NewUtils()
	rdb := redisclient.New("localhost:6379", "", 0)

	app := &App{
		HandlerUtils: handlerUtils,
		RedisClient:  rdb,
	}

	app.loadRoutes()

	return app
}

func (app *App) Start(ctx context.Context) error {

	server := &http.Server{
		Addr:    ":3001",
		Handler: app.router,
	}

	chServer := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()

		if err != nil {
			chServer <- fmt.Errorf("start http server error: %w", err)
		}

		close(chServer)

		fmt.Println("started http server")
	}()

	select {
	case err := <-chServer:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return server.Shutdown(timeout)
	}

}
