package nsenter

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/pabateman/kubectl-nsenter/internal/config"

	"github.com/pkg/errors"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

func Nsenter(clictx *cli.Context) error {
	cfg, err := config.NewConfig(clictx)
	if err != nil {
		return err
	}

	containerInfo, err := GetContainerInfo(cfg)
	if err != nil {
		return fmt.Errorf("can't get container info: %v", err)
	}

	sshConfig := &ssh.ClientConfig{
		User: cfg.SshUser,
		//HostKeyCallback: hostKeyCallback,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	if cfg.SshRequirePassword {
		password, err := requestPassword(cfg.SshUser, containerInfo.NodeIP)
		if err != nil {
			return errors.New("failed to request password")
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(password)}
	} else {
		agentConnection, err := net.Dial("unix", cfg.SshSocketPath)
		if err != nil {
			return err
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeysCallback(agent.NewClient(agentConnection).Signers)}
	}

	if cfg.SshHost != "" {
		containerInfo.NodeIP = cfg.SshHost
	}

	sshHost := net.JoinHostPort(containerInfo.NodeIP, cfg.SshPort)

	sshClient, err := ssh.Dial("tcp", sshHost, sshConfig)
	if err != nil {
		return errors.WithMessagef(err, "can't dial node %s@%s\n", cfg.SshUser, containerInfo.NodeIP)
	}

	sshSession, err := sshClient.NewSession()
	if err != nil {
		return errors.Wrap(err, "can't build ssh session")
	}
	// nolint:errcheck
	defer sshSession.Close()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return errors.Wrap(err, "failed to make tty")
	}
	// nolint:errcheck
	defer term.Restore(int(os.Stdin.Fd()), oldState)

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

	pidDiscoverCommand, err := getPidDiscoverCommand(containerInfo)
	if err != nil {
		return err
	}

	sshSession.Stdout = os.Stdout
	sshSession.Stderr = os.Stderr
	sshSession.Stdin = os.Stdin

	nsenterCommand := fmt.Sprintf("sudo nsenter -%s -t $(%s) %s",
		strings.Join(cfg.LinuxNs, " -"),
		pidDiscoverCommand,
		strings.Join(cfg.Command, " "))

	err = sshSession.Run(nsenterCommand)
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

func getPidDiscoverCommand(c *ContainerInfo) (string, error) {
	switch c.ContainerRuntime {
	case "docker":
		return fmt.Sprintf("sudo docker inspect %s --format {{.State.Pid}}", c.ContainerID), nil
	case "containerd", "cri-o":
		return fmt.Sprintf("sudo crictl inspect --output go-template --template={{.info.pid}} %s", c.ContainerID), nil
	default:
		return "", fmt.Errorf("unsupported container runtime")
	}
}
