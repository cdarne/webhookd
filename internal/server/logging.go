package server

import (
	"log"
	"net/http"
)

func Logging(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("Got request: %+v\n", r)
		next.ServeHTTP(w, r)
		logger.Printf("Sent headers: %+v\n", w.Header())
	})
}
