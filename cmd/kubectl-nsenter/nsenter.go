package main

import (
	"github.com/pabateman/kubectl-nsenter/internal/containerinfo"
	"github.com/pabateman/kubectl-nsenter/internal/sshconnect"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"strings"
)

func nsenter(clictx *cli.Context) error {
	container := clictx.String("container")
	kubeconfigPath := clictx.String("kubeconfig")
	contextOverride := clictx.String("context")
	namespaceOverride := clictx.String("namespace")
	podName := clictx.Args().First()
	command := clictx.Args().Tail()
	if len(command) == 0 {
		return errors.Wrap(cli.ShowAppHelp(clictx), "you must provide a command")
	}
	kubeconfigFiles := strings.Split(kubeconfigPath, ":")
	containerInfo, err := containerinfo.GetContainerInfo(
		kubeconfigFiles,
		contextOverride,
		namespaceOverride,
		podName,
		container)
	if err != nil {
		return errors.Wrap(err, "can't get container info")
	}

	sshClient, err := sshconnect.BuildSshClient(containerInfo.NodeName, containerInfo.NodeIP)
	if err != nil {
		return errors.Wrap(err, "can't build ssh client")
	}
	defer sshClient.Close()
	sshconnect.SshExecute(sshClient, strings.Join(command, " "))
	return nil
}
