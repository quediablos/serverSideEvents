package handler

import (
	"context"
	"fmt"
	"net/http"
	"serverSideEvents/client/redis"
	"time"
)

type PriceTickerSseHandler struct {
	Utils       *Utils
	RedisClient *redis.Client
}

func (h *PriceTickerSseHandler) Get(w http.ResponseWriter, r *http.Request) {

	currency := r.URL.Query().Get("currency")
	if currency == "" {
		http.Error(w, "missing required query parameter: currency", http.StatusBadRequest)
		return
	}

	h.Utils.SetResponseHeaders(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	fmt.Println("Client connected")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Client disconnected")
			return

		case <-ticker.C:
			price, err := h.GetTicker(ctx, currency)
			if err != nil {
				fmt.Fprintf(w, "data: error: %s\n\n", err.Error())
				flusher.Flush()
				return
			}

			fmt.Fprintf(w, "data: Symbol:%s Price:%s\n\n", currency, price)
			flusher.Flush()
		}
	}
}

func (h *PriceTickerSseHandler) GetTicker(ctx context.Context, currency string) (string, error) {
	price, err := h.RedisClient.Get(ctx, currency, "price")
	if err != nil {
		return "", err
	}

	return price, nil
}
