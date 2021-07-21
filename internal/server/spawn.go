package server

import (
	"fmt"
	"github.com/cdarne/webhookd/internal/subprocess"
	"io"
	"net/http"
)

func SpawnProcess(command string, commandArgs []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error while reading the request body: %s", err), http.StatusBadRequest)
			return
		}
		go subprocess.Run(r, command, commandArgs, body)

		w.WriteHeader(http.StatusNoContent)
	})
}
