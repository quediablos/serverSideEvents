package handler

import (
	"fmt"
	"net/http"
	"time"
)

type ExampleSseHandler struct {
}

func (*ExampleSseHandler) Post(w http.ResponseWriter, r *http.Request) {
	// 1. Set SSE-specific HTTP headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // For development CORS

	// 2. Ensure the ResponseWriter supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// 3. Monitor connection closure using request context
	ctx := r.Context()

	fmt.Println("Client connected")

	// 4. Run the event loop
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Client disconnected or request canceled
			fmt.Println("Client disconnected")
			return

		case t := <-ticker.C:
			// Formulate message payload according to the spec
			// Format must end with two newlines \n\n
			fmt.Fprintf(w, "data: The current time is %s\n\n", t.Format(time.RFC3339))

			// Flush data immediately to the client
			flusher.Flush()
		}
	}
}
