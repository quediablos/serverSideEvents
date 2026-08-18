package app

import (
	"serverSideEvents/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *App) loadRoutes() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Route("/", app.loadOrderRoutes)
	app.router = router
}

func (app *App) loadOrderRoutes(router chi.Router) {
	sseHandler := &handler.ExampleSseHandler{}

	router.Post("/example-sse-handler", sseHandler.Post)
}
