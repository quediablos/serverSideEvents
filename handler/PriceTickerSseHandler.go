package handler

import (
	"context"
	"fmt"
	"net/http"
	"serverSideEvents/client/redis"
	"serverSideEvents/timing"
	"strings"
	"time"
)

type PriceTickerSseHandler struct {
	Utils       *Utils
	RedisClient *redis.Client
	TimingUtils *timing.Utils
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

	//Sleep until 80 millisecond mark of the next second.
	h.TimingUtils.Synchronize(ctx, 80)

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

	var price string
	timestamp := time.Now().UTC().Format("2006_01_02_15_04_05")

	priceWithTimestamp, err := h.RedisClient.Get(ctx, fmt.Sprintf("%s:%s", currency, timestamp), "price")

	//General error, not indicating the data not found.
	if err != nil && !strings.Contains(err.Error(), timestamp) {
		return "", err
	}

	//Get the last ticker, the one without timestamp.
	if priceWithTimestamp == "" {
		priceWithoutTimestamp, err := h.RedisClient.Get(ctx, currency, "price")

		if err != nil && !strings.Contains(err.Error(), timestamp) {
			return "", err
		}

		price = priceWithoutTimestamp
	} else {
		price = priceWithTimestamp
	}

	return price, nil
}
