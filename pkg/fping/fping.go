package fping

import (
	"bufio"
	"log"
	"os/exec"
	"syscall"
)

type FpingProcess struct {
	Responses chan *Response
	cmd       *exec.Cmd
	stopping  bool
}

func NewFpingProcess(network string) *FpingProcess {
	return &FpingProcess{
		Responses: make(chan *Response),
		// Long options were introduced in fping 4
		// -A, --addr
		// -e, --elapsed
		// -l, --loop
		// -g, --generate addr/mask
		cmd:      exec.Command("fping", "-A", "-e", "-l", "-g", network),
		stopping: false,
	}
}

func (fping *FpingProcess) Start() error {
	stdout, err := fping.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := fping.cmd.Start(); err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			resp := Parseline(line)
			fping.Responses <- resp
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}()
	return nil
}

func (fping *FpingProcess) Stop() {
	fping.cmd.Process.Signal(syscall.SIGTERM)
	fping.cmd.Process.Wait()
}
