package main

import (
	"fmt"
	"github.com/pabateman/kubectl-nsenter/internal/containerinfo"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"strings"
)

func nsenter(clictx *cli.Context) error {
	container := clictx.String("container")
	kubeconfigPath := clictx.String("kubeconfig")
	podName := clictx.Args().First()
	command := clictx.Args().Tail()
	if len(command) == 0 {
		return errors.Wrap(cli.ShowAppHelp(clictx), "you must provide a command")
	}
	kubeconfigFiles := strings.Split(kubeconfigPath, ":")
	containerInfo, err := containerinfo.GetContainerInfo(kubeconfigFiles, podName, container)
	if err != nil {
		return errors.Wrap(err, "can't get container info")
	}

	return fmt.Errorf("Command is %s. Node Name is %s. Container ID is  %s", command, containerInfo.NodeName, containerInfo.ContainerID)
}
