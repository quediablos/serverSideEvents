package app

import (
	"context"
	"fmt"
	"net/http"
	redisclient "serverSideEvents/client/redis"
	"serverSideEvents/handler"
	"serverSideEvents/job"
	"serverSideEvents/timing"
	"time"
)

type App struct {
	router           http.Handler
	HandlerUtils     *handler.Utils
	RedisClient      *redisclient.Client
	PriceSeederJob   *job.PriceSeederJob
	CandleBuilderJob *job.CandleBuilderJob
	TimingUtils      *timing.Utils
}

func New() *App {

	handlerUtils := handler.NewUtils()
	rdb := redisclient.New("localhost:6379", "", 0)
	timingUtils := timing.New()
	priceSeederJob := job.NewPriceSeederJob(rdb, timingUtils, "USDTRY")
	candleBuilderJob := job.NewCandleBuilderJob(rdb, timingUtils, "USDTRY", job.Period1Min)

	app := &App{
		HandlerUtils:     handlerUtils,
		RedisClient:      rdb,
		PriceSeederJob:   priceSeederJob,
		CandleBuilderJob: candleBuilderJob,
		TimingUtils:      timingUtils,
	}

	app.loadRoutes()

	return app
}

func (app *App) Start(ctx context.Context) error {

	go app.PriceSeederJob.Run(ctx)
	go app.CandleBuilderJob.Run(ctx)

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
