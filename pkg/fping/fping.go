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

func NewFpingProcess() *FpingProcess {
	return &FpingProcess{
		Responses: make(chan *Response),
		cmd:       exec.Command("fping", "--addr", "--elapsed", "--loop", "--generate", "192.168.1.0/24"),
		stopping:  false,
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
