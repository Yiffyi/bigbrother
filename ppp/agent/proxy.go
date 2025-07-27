package agent

import (
	"bytes"
	"encoding/gob"
	"errors"
	"os/exec"

	"github.com/rs/zerolog/log"
	"github.com/yiffyi/bigbrother/ppp/model"
	"golang.org/x/crypto/ssh"
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

func (p *Proxy) HandleSSHRequest(sshReq *ssh.Request) error {
	var err error

	reqDec := gob.NewDecoder(bytes.NewReader(sshReq.Payload))

	replyBuffer := bytes.NewBuffer([]byte{})
	replyEnc := gob.NewEncoder(replyBuffer)
	switch sshReq.Type {
	case "updateProxyConfig":
		defer func() {
			if sshReq.WantReply {
				replyEnc.Encode(model.UpdateProxyConfigResponse{
					Error: err,
				})
				sshReq.Reply(err == nil, replyBuffer.Bytes())
			}
		}()

		var req model.UpdateProxyConfigRequest
		err = reqDec.Decode(&req)
		if err != nil {
			return err
		}

		err = p.UpdateProxyConfig(req.ConfigFile, req.Restart)
		// if err != nil {
		// 	return err
		// }

		return err
	default:
		if sshReq.WantReply {
			return sshReq.Reply(false, nil)
		} else {
			return errors.New("unsupported ssh request")
		}
	}
}
