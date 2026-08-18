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
			price, err := h.GetTicker(ctx)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				flusher.Flush()
				return
			}

			fmt.Fprintf(w, "data: Symbol:%s Price:%s\n\n", "USD", price) //TODO: parametrize the currency
			flusher.Flush()
		}
	}
}

func (h *PriceTickerSseHandler) GetTicker(ctx context.Context) (string, error) {

	//TODO:implement symbol parameter
	price, err := h.RedisClient.Get(ctx, "USD")
	if err != nil {
		return "", err
	}

	return price, nil
}
