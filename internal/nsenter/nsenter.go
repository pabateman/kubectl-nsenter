package nsenter

import (
	"fmt"
	"strings"

	"github.com/pabateman/kubectl-nsenter/internal/config"
	"github.com/pabateman/kubectl-nsenter/internal/containerinfo"
	"github.com/pabateman/kubectl-nsenter/internal/k8s"
	"github.com/pabateman/kubectl-nsenter/internal/ssh"

	cli "github.com/urfave/cli/v2"
)

func Nsenter(clictx *cli.Context) error {
	cfg, err := config.NewConfig(clictx)
	if err != nil {
		return err
	}

	kubeClient, err := k8s.GetKubernetesClient(cfg)
	if err != nil {
		return fmt.Errorf("can't build client: %v", err)
	}

	pod, err := k8s.GetPod(cfg, kubeClient)
	if err != nil {
		return fmt.Errorf("can't get pod spec: %v", err)
	}

	containerInfo, err := containerinfo.GetContainerInfo(cfg, pod)
	if err != nil {
		return fmt.Errorf("can't get container info: %v", err)
	}

	if cfg.SSHHost == "" {
		cfg.SSHHost = containerInfo.NodeIP
	}

	pidDiscoverCommand, err := containerinfo.GetPidDiscoverCommand(containerInfo)
	if err != nil {
		return err
	}

	nsenterCommand := fmt.Sprintf("sudo nsenter -%s -t $(%s) %s",
		strings.Join(cfg.LinuxNs, " -"),
		pidDiscoverCommand,
		strings.Join(cfg.Command, " "))

	err = ssh.ExecSSHCommand(cfg, nsenterCommand)
	if err != nil {
		return fmt.Errorf("failed to execute ssh command: %v", err)
	}

	return nil
}
