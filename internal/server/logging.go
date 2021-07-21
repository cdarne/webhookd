package server

import (
	"log"
	"net/http"
)

func Logging(logger *log.Logger, next handlerWithError) handlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		logger.Printf("Got request: %+v\n", r)
		status := 204
		var body string
		err := next(w, r)
		if err != nil {
			body = err.Error()
			clientError, ok := err.(ClientError)
			if ok {
				status = clientError.Status()
			} else {
				status = 500
			}
		}
		logger.Printf("[%d]: %s\n", status, body)
		return err
	}
}
