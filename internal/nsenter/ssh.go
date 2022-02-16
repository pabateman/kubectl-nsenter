package nsenter

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"os"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

func DetermineSSHAuth(sshConfig *ssh.ClientConfig, host string) (*ssh.Client, error) {
	var client = new(ssh.Client)
	var authSock = os.Getenv("SSH_AUTH_SOCK")
	var agentConnection, err = net.Dial("unix", authSock)
	if err == nil {
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeysCallback(agent.NewClient(agentConnection).Signers)}
		client, err = ssh.Dial("tcp", host, sshConfig)
		if err == nil {
			return client, nil
		}
	}

	var password string
	fmt.Print("Enter SSH password:")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read password")
	}
	fmt.Println()
	password = string(bytePassword)
	sshConfig.Auth = []ssh.AuthMethod{ssh.Password(password)}
	client, err = ssh.Dial("tcp", host, sshConfig)
	if err == nil {
		return client, nil
	}
	client.Close()
	return nil, errors.New("failed to connect to host via ssh")
}

func BuildSshClient(ip string, user string) (*ssh.Client, error) {

	sshAddr := ip
	sshHost := fmt.Sprintf("%s:%s", sshAddr, "22")

	//knownHostsFilePath := path.Join(getSystemUser().HomeDir, sshHome, "known_hosts")
	//hostKeyCallback, err := knownhosts.New(knownHostsFilePath)
	//if err != nil {
	//	return nil, errors.Wrap(err, "can't get ssh known hosts")
	//}
	sshConfig := &ssh.ClientConfig{
		User: user,
		//HostKeyCallback: hostKeyCallback,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := DetermineSSHAuth(sshConfig, sshHost)
	if err != nil {
		return nil, errors.Wrap(err, "can't build ssh client")
	}
	return client, nil
}

func SshExecute(client *ssh.Client, command string) error {
	session, err := client.NewSession()
	if err != nil {
		return errors.Wrap(err, "failed to connect to node")
	}
	defer session.Close()
	var stdout bytes.Buffer
	session.Stdout = &stdout
	session.Run(command)
	fmt.Println(stdout.String())
	return nil
}
