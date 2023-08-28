package nsenter

import (
	"fmt"
	"strings"

	"github.com/pabateman/kubectl-nsenter/internal/config"
	"github.com/pabateman/kubectl-nsenter/internal/containerinfo"
	"github.com/pabateman/kubectl-nsenter/internal/k8s"
	"github.com/pabateman/kubectl-nsenter/internal/util"

	cli "github.com/urfave/cli/v2"
)

const (
	sshCmd     = "ssh"
	nsenterCmd = "nsenter"
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

	if cfg.UseNodeName {
		cfg.SSHHost = containerInfo.NodeName
	}

	cmd := buildSSHCmd(cfg)
	nsenterCmd, err := buildNsenterCmd(cfg, containerInfo)
	if err != nil {
		return err
	}

	cmd = append(cmd, strings.Join(nsenterCmd, " "))

	err = util.Fork(cmd, cfg.Interactive)
	if err != nil {
		return fmt.Errorf("failed to execute ssh command: %v", err)
	}

	return nil
}

func buildSSHCmd(cfg *config.Config) []string {
	result := make([]string, 0)

	result = append(result, sshCmd)

	if !cfg.TTY {
		result = append(result, "-T")
	} else {
		result = append(result, "-t")
	}

	if cfg.SSHPort != "" {
		result = append(result, []string{"-p", cfg.SSHPort}...)
	}

	for _, opt := range cfg.SSHOpts {
		result = append(result, []string{"-o", opt}...)
	}

	if cfg.SSHUser != "" {
		result = append(result, fmt.Sprintf(
			"%s@%s",
			cfg.SSHUser,
			cfg.SSHHost,
		))
	} else {
		result = append(result, cfg.SSHHost)
	}

	return result
}

func buildNsenterCmd(cfg *config.Config, containerInfo *containerinfo.ContainerInfo) ([]string, error) {
	result := make([]string, 0)

	result = append(result, []string{"sudo", nsenterCmd}...)

	for _, ns := range cfg.LinuxNs {
		result = append(result, fmt.Sprintf("-%s", ns))
	}

	result = append(result, "-t")

	pidDiscoverCmd, err := containerinfo.GetPidDiscoverCommand(containerInfo)
	if err != nil {
		return nil, err
	}

	pidDiscoverCmd[0] = fmt.Sprintf("$(%s", pidDiscoverCmd[0])
	pidDiscoverCmd[len(pidDiscoverCmd)-1] = fmt.Sprintf("%s)", pidDiscoverCmd[len(pidDiscoverCmd)-1])

	result = append(result, pidDiscoverCmd...)
	result = append(result, cfg.Command...)

	return result, nil
}
