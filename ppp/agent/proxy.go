package agent

import (
	"bytes"
	"errors"
	"os/exec"

	"github.com/rs/zerolog/log"
)

type Proxy struct {
	DaemonType string
	Config     []byte

	process *exec.Cmd
	program string
	args    []string
}

func NewProxy(daemonType string, program string, args []string, config []byte) (*Proxy, error) {
	if _, err := exec.LookPath(program); err != nil {
		return nil, err
	}

	return &Proxy{
		DaemonType: daemonType,
		Config:     config,

		process: nil,
		program: program,
		args:    args,
	}, nil
}

func (p *Proxy) monitor() {
	err := p.process.Wait()
	if err != nil {
		log.Warn().Err(err).Msg("proxy process stopped with error")
	} else {
		log.Info().Int("code", p.process.ProcessState.ExitCode()).Msg("proxy process exited normally")
	}
	p.process = nil
}

func (p *Proxy) Start() error {
	if p.process != nil {
		return errors.New("proxy process already started")
	}

	p.process = exec.Command(p.program, p.args...)

	if p.Config != nil {
		p.process.Stdin = bytes.NewReader(p.Config)
	}

	err := p.process.Start()
	if err != nil {
		p.process = nil
		return err
	}

	go p.monitor()
	return nil
}

func (p *Proxy) Stop() error {
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

func (p *Proxy) IsRunning() bool {
	if p.process == nil {
		return false
	}

	if p.process.Process == nil {
		return false
	}

	return true
}

func (p *Proxy) UpdateProxyConfig(config []byte, restart bool) error {
	p.Config = config

	if restart {
		p.Stop()
		return p.Start()
	}
	return nil
}
