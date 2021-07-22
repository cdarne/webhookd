package subprocess

import (
	"log"
	"os/exec"
	"sync"
)

type Runner struct {
	concurrency int
	logger      *log.Logger
	commands    chan *exec.Cmd
	wg          *sync.WaitGroup
}

func NewRunner(logger *log.Logger, concurrency, queueSize int) *Runner {
	return &Runner{
		concurrency: concurrency,
		logger:      logger,
		commands:    make(chan *exec.Cmd, queueSize),
		wg:          new(sync.WaitGroup),
	}
}

func (w *Runner) Stop() {
	close(w.commands)
	w.wg.Wait()
}

func (w *Runner) Start() {
	for i := 0; i < w.concurrency; i++ {
		w.wg.Add(1)
		go runner(w.commands, w.logger, w.wg)
	}
}

func runner(cmds <-chan *exec.Cmd, logger *log.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	var err error
	for cmd := range cmds {
		err = cmd.Run()
		if err != nil {
			logger.Printf("Error while running the %s command: %s\n", cmd, err)
		}
	}
}

func (w *Runner) Enqueue(cmd *exec.Cmd) bool {
	select {
	case w.commands <- cmd:
		return true
	default:
		return false
	}
}
