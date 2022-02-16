package nsenter

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
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

	sshClient, err := BuildSshClient(containerInfo.NodeIP, sshUser)
	if err != nil {
		return fmt.Errorf("can't build ssh client: %v", err)
	}
	defer sshClient.Close()
	SshExecute(sshClient, strings.Join(command, " "))
	return nil
}
