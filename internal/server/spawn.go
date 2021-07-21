package server

import (
	"fmt"
	"github.com/cdarne/webhookd/internal/subprocess"
	"io"
	"net/http"
)

func SpawnProcess(command string, commandArgs []string) handlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("error while reading the request body: %s", err)
		}
		go subprocess.Run(r, command, commandArgs, body)

		return nil
	}
}
