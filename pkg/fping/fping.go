package fping

import (
	"bufio"
	"io"
	"log"
	"os/exec"
	"syscall"
)

type FpingProcess struct {
	Responses    chan *Response
	Unreachables chan *UnreachableResponse
	cmd          *exec.Cmd
	stopping     bool
}

func NewFpingProcess(network string) *FpingProcess {
	return &FpingProcess{
		Responses:    make(chan *Response),
		Unreachables: make(chan *UnreachableResponse),
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
	stderr, err := fping.cmd.StderrPipe()
	if err != nil {
		return err
	}
	go func() {
		if err := fping.cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	go fping.handleStdout(stdout)
	go fping.handleStderr(stderr)

	return nil
}

func (fping *FpingProcess) handleStdout(stdout io.Reader) {
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		resp := Parseline(line)
		fping.Responses <- resp
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (fping *FpingProcess) handleStderr(stderr io.Reader) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		resp := ParseStderr(line)
		if resp == nil {
			// fmt.Fprintf(os.Stderr, "Failed to parse line: '%s'\n", line)
			continue
		}
		fping.Unreachables <- resp
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (fping *FpingProcess) Stop() {
	fping.cmd.Process.Signal(syscall.SIGTERM)
	fping.cmd.Process.Wait()
}
