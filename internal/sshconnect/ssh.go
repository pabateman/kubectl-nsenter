package sshconnect

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"os/user"
	"path"
	"syscall"
	"time"

	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
	//"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

const (
	sshHome = ".ssh"
)

func getSystemUser() *user.User {
	systemUser, _ := user.Current()
	return systemUser
}

func getPrivateKeys(sshhome string) ([]ssh.Signer, error) {
	sshFiles, err := ioutil.ReadDir(sshhome)
	if err != nil {
		return nil, errors.Wrap(err, "can't reach ssh home (~/.ssh) directory")
	}
	var signers = make([]ssh.Signer, 0)
	for _, file := range sshFiles {
		if file.IsDir() {
			continue
		}
		filepath := path.Join(sshhome, file.Name())
		key, err := ioutil.ReadFile(filepath)
		if err != nil {
			return nil, errors.Wrapf(err, "can't access %s", filepath)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err == nil {
			signers = append(signers, signer)
		}
	}
	return signers, nil
}

func DetermineSSHAuth(sshConfig *ssh.ClientConfig, host string) (*ssh.Client, error) {
	sshHomeDir := path.Join(getSystemUser().HomeDir, sshHome)
	privateKeys, err := getPrivateKeys(sshHomeDir)

	var client = new(ssh.Client)
	for _, key := range privateKeys {
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(key)}
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

func BuildSshClient(alias string, ip string) (*ssh.Client, error) {

	sshAddr := ssh_config.Get(alias, "HostName")
	if sshAddr == "" {
		sshAddr = ip
	}
	sshUser := ssh_config.Get(alias, "User")
	if sshUser == "" {
		sshUser = getSystemUser().Username
	}
	sshPort := ssh_config.Get(alias, "Port")
	sshHost := fmt.Sprintf("%s:%s", sshAddr, sshPort)

	//knownHostsFilePath := path.Join(getSystemUser().HomeDir, sshHome, "known_hosts")
	//hostKeyCallback, err := knownhosts.New(knownHostsFilePath)
	//if err != nil {
	//	return nil, errors.Wrap(err, "can't get ssh known hosts")
	//}
	sshConfig := &ssh.ClientConfig{
		User: sshUser,
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
