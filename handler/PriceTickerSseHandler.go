package handler

import (
	"fmt"
	"net/http"
	"time"
)

type PriceTickerSseHandler struct {
	Utils *Utils
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

		case t := <-ticker.C:
			fmt.Fprintf(w, "data: The current time is %s\n\n", t.Format(time.RFC3339))
			flusher.Flush()
		}
	}
}
