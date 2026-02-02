package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Engine struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
}

func NewEngine(path string) (*Engine, error) {
	cmd := exec.Command(path)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	engine := &Engine{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdoutPipe),
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	engine.send("uci")
	engine.waitFor("uciok")

	engine.send("isready")
	engine.waitFor("readyok")

	return engine, nil
}

func (e *Engine) send(cmd string) {
	fmt.Fprintln(e.stdin, cmd)
}

func (e *Engine) waitFor(token string) {
	for {
		line, _ := e.stdout.ReadString('\n')
		if strings.Contains(line, token) {
			return
		}
	}
}

func (e *Engine) SetElo(elo int) {
	e.send("setoption name UCI_LimitStrength value true")
	e.send(fmt.Sprintf("setoption name UCI_Elo value %d", elo))
}

func (e *Engine) BestMove(fen string, thinkTimeMs int) (string, error) {
	e.send("position fen " + fen)
	e.send(fmt.Sprintf("go movetime %d", thinkTimeMs))

	for {
		line, _ := e.stdout.ReadString('\n')
		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Split(line, " ")
			return parts[1], nil
		}
	}
}
