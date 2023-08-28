// Deprecated package in favour of simple fork ssh client on host
package ssh

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/pabateman/kubectl-nsenter/internal/config"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

func ExecSSHCommand(cfg *config.Config, cmd string) error {
	sshSession, err := getSSHSession(cfg)
	if err != nil {
		return fmt.Errorf("failed to build ssh session: %v", err)
	}
	// nolint:errcheck
	defer sshSession.Close()

	previousState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return errors.Wrap(err, "failed to make tty")
	}
	// nolint:errcheck
	defer term.Restore(int(os.Stdin.Fd()), previousState)

	ttyWidth, ttyHeight, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return errors.Wrap(err, "failed to get tty size")
	}

	modes := ssh.TerminalModes{
		ssh.ECHO: 1,
	}

	err = sshSession.RequestPty("xterm", ttyHeight, ttyWidth, modes)
	if err != nil {
		return errors.Wrap(err, "failed to request pty")
	}

	sshSession.Stdout = os.Stdout
	sshSession.Stderr = os.Stderr
	sshSession.Stdin = os.Stdin

	err = sshSession.Run(cmd)
	if err != nil {
		return fmt.Errorf("remote shell exited with non zero code: %v", err)
	}

	return nil
}

func requestPassword(user, host string) (string, error) {
	fmt.Printf("%s@%s's password: ", user, host)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", errors.New("failed to read password")
	}
	return string(password), nil
}

func getSSHSession(cfg *config.Config) (*ssh.Session, error) {
	sshConfig := &ssh.ClientConfig{
		User:            cfg.SSHUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	if cfg.SSHRequirePassword {
		password, err := requestPassword(cfg.SSHUser, cfg.SSHHost)
		if err != nil {
			return nil, errors.New("failed to request password")
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(password)}
	} else {
		agentConnection, err := net.Dial("unix", cfg.SSHSocketPath)
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeysCallback(agent.NewClient(agentConnection).Signers)}
	}

	sshHost := net.JoinHostPort(cfg.SSHHost, cfg.SSHPort)

	sshClient, err := ssh.Dial("tcp", sshHost, sshConfig)
	if err != nil {
		return nil, errors.WithMessagef(err, "can't dial node %s@%s\n", cfg.SSHUser, cfg.SSHHost)
	}

	return sshClient.NewSession()
}
