package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/cdarne/webhookd/internal/subprocess"
)

func SpawnProcess(command string, commandArgs []string, runner *subprocess.Runner) handlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("error while reading the request body: %s", err)
		}
		if !runner.Enqueue(subprocess.NewCommand(r, command, commandArgs, body)) {
			return NewHTTPError(http.StatusTooManyRequests, errors.New("runner queue is full"))
		}
		return nil
	}
}
