package nsenter

import (
	"fmt"
	"net"
	"os"

	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

func requestPassword(user, host string) (string, error) {
	fmt.Printf("%s@%s's password: ", user, host)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", errors.New("failed to read password")
	}
	return string(password), nil
}

func Nsenter(clictx *cli.Context) error {
	podName := clictx.Args().First()
	if podName == "" {
		fmt.Println("you must specify pod name!")
		return cli.ShowAppHelp(clictx)
	}

	command := clictx.Args().Tail()
	if len(command) == 0 {
		fmt.Println("you must provide a command!")
		return cli.ShowAppHelp(clictx)
	}

	containerInfo, err := GetContainerInfo(clictx)
	if err != nil {
		return errors.WithMessage(err, "can't get container info")
	}

	sshUser := clictx.String("user")
	sshAuthSock := clictx.String("ssh-auth-sock")

	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		//HostKeyCallback: hostKeyCallback,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	if clictx.Bool("password") {
		password, err := requestPassword(sshUser, containerInfo.NodeIP)
		if err != nil {
			return errors.New("failed to request password")
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(password)}
	} else {
		agentConnection, err := net.Dial("unix", sshAuthSock)
		if err == nil {
			sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeysCallback(agent.NewClient(agentConnection).Signers)}
		}
	}

	if clictx.String("host") != "" {
		containerInfo.NodeIP = clictx.String("host")
	}

	sshPort := clictx.String("port")
	sshHost := net.JoinHostPort(containerInfo.NodeIP, sshPort)

	sshClient, err := ssh.Dial("tcp", sshHost, sshConfig)
	if err != nil {
		return errors.WithMessagef(err, "can't dial node %s@%s\n", sshUser, containerInfo.NodeIP)
	}

	sshSession, err := sshClient.NewSession()
	if err != nil {
		return errors.Wrap(err, "can't build ssh session")
	}
	defer sshSession.Close()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return errors.Wrap(err, "failed to make tty")
	}
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

	var pidDiscoverCommand string
	switch containerInfo.ContainerRuntime {
	case "docker":
		pidDiscoverCommand = fmt.Sprintf("sudo docker inspect %s --format {{.State.Pid}}", containerInfo.ContainerID)
	case "containerd", "cri-o":
		pidDiscoverCommand = fmt.Sprintf("sudo crictl inspect --output go-template --template={{.info.pid}} %s", containerInfo.ContainerID)
	default:
		return fmt.Errorf("unsupported container runtime")
	}

	sshSession.Stdout = os.Stdout
	sshSession.Stderr = os.Stderr
	sshSession.Stdin = os.Stdin

	nsenterNamespaces := clictx.StringSlice("ns")
	nsenterCommand := fmt.Sprintf("sudo nsenter -%s -t $(%s) %s",
		strings.Join(nsenterNamespaces, " -"),
		pidDiscoverCommand,
		strings.Join(command, " "))

	err = sshSession.Run(nsenterCommand)
	if err != nil {
		return errors.Wrap(err, "failed to start shell: %s")
	}

	return nil
}
