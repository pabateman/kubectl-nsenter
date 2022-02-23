package nsenter

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func Nsenter(clictx *cli.Context) error {
	container := clictx.String("container")
	kubeconfigPath := clictx.String("kubeconfig")
	contextOverride := clictx.String("context")
	namespaceOverride := clictx.String("namespace")
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
	kubeconfigFiles := strings.Split(kubeconfigPath, ":")
	containerInfo, err := GetContainerInfo(
		kubeconfigFiles,
		contextOverride,
		namespaceOverride,
		podName,
		container)
	if err != nil {
		return fmt.Errorf("can't get container info: %v", err)
	}

	sshUser := clictx.String("user")
	sshAuthSock := clictx.String("ssh-auth-sock")
	sshPort := clictx.String("port")
	sshHost := net.JoinHostPort(containerInfo.NodeIP, sshPort)

	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		//HostKeyCallback: hostKeyCallback,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	sshClient := new(ssh.Client)

	agentConnection, err := net.Dial("unix", sshAuthSock)
	if err == nil {
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeysCallback(agent.NewClient(agentConnection).Signers)}
		sshClient, err = ssh.Dial("tcp", sshHost, sshConfig)
		if err != nil {
			return errors.Wrapf(err, "can't dial node %s@%s", sshUser, sshHost)
		}
	}

	sshSession, err := sshClient.NewSession()
	if err != nil {
		return errors.Wrap(err, "can't build ssh session")
	}
	defer sshSession.Close()
	var stdout bytes.Buffer

	sudoCheckCommand := "sudo true; echo $?"

	sshSession.Stdout = &stdout
	sshSession.Run(sudoCheckCommand)
	fmt.Println(stdout.String())
	if stdout.String() != "0" {
		return fmt.Errorf("sudo requiried")
	}

	return nil
}
