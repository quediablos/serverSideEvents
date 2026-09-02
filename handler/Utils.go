package handler

import (
	"net/http"
)

type Utils struct{}

func NewUtils() *Utils {
	return &Utils{}
}

func (*Utils) SetResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}
