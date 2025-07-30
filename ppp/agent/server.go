package agent

import (
	"bytes"
	"errors"
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
	"github.com/yiffyi/bigbrother/ppp/model"
)

type Server struct {
	ServerType model.ProgramType
	Config     []byte

	process      *exec.Cmd
	program      string
	args         []string
	shareConsole bool
}

func NewServer(serverType model.ProgramType, program string, args []string, config []byte, shareConsole bool) (*Server, error) {
	if _, err := exec.LookPath(program); err != nil {
		return nil, err
	}

	return &Server{
		ServerType: serverType,
		Config:     config,

		process:      nil,
		program:      program,
		args:         args,
		shareConsole: shareConsole,
	}, nil
}

func (p *Server) monitor() {
	err := p.process.Wait()
	if err != nil {
		log.Warn().Err(err).Msg("proxy process stopped with error")
	} else {
		log.Info().Int("code", p.process.ProcessState.ExitCode()).Msg("proxy process exited normally")
	}
	p.process = nil
}

func (p *Server) Start() error {
	if p.process != nil {
		return errors.New("proxy process already started")
	}

	p.process = exec.Command(p.program, p.args...)

	if p.Config != nil {
		p.process.Stdin = bytes.NewReader(p.Config)
	}

	if p.shareConsole {
		p.process.Stdout = os.Stdout
		p.process.Stderr = os.Stderr
	}

	err := p.process.Start()
	if err != nil {
		p.process = nil
		return err
	}

	go p.monitor()
	return nil
}

func (p *Server) Stop() error {
	if p.process == nil {
		return errors.New("proxy process not started")
	}

	if p.process.Process == nil {
		p.process = nil
		return errors.New("proxy process created, but not started")
	}

	err := p.process.Process.Kill()
	if err != nil {
		return err
	}

	p.process = nil
	return nil
}

func (p *Server) IsRunning() bool {
	if p.process == nil {
		return false
	}

	if p.process.Process == nil {
		return false
	}

	return true
}

func (p *Server) UpdateServerConfig(config []byte, restart bool) error {
	p.Config = config

	if restart {
		p.Stop()
		return p.Start()
	}
	return nil
}
