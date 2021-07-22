package subprocess

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

var dashToUnderscore = strings.NewReplacer("-", "_")
var newlineToSpace = strings.NewReplacer("\r", " ", "\n", " ")

func NewCommand(req *http.Request, commandName string, commandArgs []string, body []byte) *exec.Cmd {
	env := setupEnv(req)
	return newCmd(commandName, commandArgs, body, env)
}

func newCmd(commandName string, commandArgs []string, body []byte, env []string) *exec.Cmd {
	cmd := exec.Command(commandName, commandArgs...)
	cmd.Env = env
	cmd.Stdin = bytes.NewReader(body)
	return cmd
}

func setupEnv(req *http.Request) (env []string) {
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
