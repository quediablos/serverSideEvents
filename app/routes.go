package app

import (
	"serverSideEvents/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *App) loadRoutes() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Route("/examples", app.loadExampleRoutes)
	router.Route("/prices", app.loadPriceTickerRoutes)
	app.router = router
}

func (app *App) loadExampleRoutes(router chi.Router) {
	examplesSseHandler := &handler.ExampleSseHandler{
		Utils: app.HandlerUtils,
	}

	router.Get("/", examplesSseHandler.Get)
}

func (app *App) loadPriceTickerRoutes(router chi.Router) {

	priceTickerHandler := &handler.PriceTickerSseHandler{
		Utils:       app.HandlerUtils,
		RedisClient: app.RedisClient,
	}

	router.Get("/", priceTickerHandler.Get)
}
