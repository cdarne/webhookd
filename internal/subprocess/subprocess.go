package subprocess

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

var dashToUnderscore = strings.NewReplacer("-", "_")
var newlineToSpace = strings.NewReplacer("\r", " ", "\n", " ")

func Run(req *http.Request, commandName string, commandArgs []string, body []byte, logger *log.Logger) {
	env := SetupEnv(req)
	err := runCmd(commandName, commandArgs, body, env)
	if err != nil {
		logger.Println("Error while running the command: ", err)
	}
}

func runCmd(commandName string, commandArgs []string, body []byte, env []string) error {
	cmd := exec.Command(commandName, commandArgs...)
	cmd.Env = env
	cmd.Stdin = bytes.NewReader(body)
	return cmd.Run()
}

func SetupEnv(req *http.Request) (env []string) {
	for k, vals := range req.Header {
		header := fmt.Sprintf("HTTP_%s", dashToUnderscore.Replace(k))
		env = appendEnv(env, header, vals...)
	}
	return env
}

func appendEnv(env []string, key string, values ...string) []string {
	if len(values) == 0 {
		return env
	}

	trimmedValues := make([]string, 0, len(values))
	for _, val := range values {
		trimmedValues = append(trimmedValues, strings.TrimSpace(newlineToSpace.Replace(val)))
	}
	return append(env, fmt.Sprintf("%s=%s",
		strings.ToUpper(key),
		strings.Join(trimmedValues, ", ")))
}
