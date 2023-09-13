package containerinfo

import (
	"fmt"
	"path"
	"strings"

	"github.com/pabateman/kubectl-nsenter/internal/config"

	v1 "k8s.io/api/core/v1"
)

type ContainerInfo struct {
	NodeName         string
	NodeIP           string
	ContainerID      string
	ContainerRuntime string
}

func GetContainerInfo(cfg *config.Config, pod *v1.Pod) (*ContainerInfo, error) {
	var containerStatus *v1.ContainerStatus
	var err error

	if !podInitialized(pod) {
		containerStatus, err = getInitializingContainerStatus(pod)
		if err != nil {
			return nil, fmt.Errorf("pod at very initializing phase: %v", err)
		}
	} else {
		containerStatus, err = getContainerStatus(pod, cfg)
		if err != nil {
			return nil, fmt.Errorf("can't get container status: %v", err)
		}
	}

	return &ContainerInfo{
		NodeName:         pod.Spec.NodeName,
		NodeIP:           pod.Status.HostIP,
		ContainerID:      path.Base(containerStatus.ContainerID),
		ContainerRuntime: strings.Split(containerStatus.ContainerID, ":")[0],
	}, nil
}

func podInitialized(pod *v1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodInitialized {
			return condition.Status == v1.ConditionTrue
		}
	}
	return false
}

func getInitializingContainerStatus(pod *v1.Pod) (*v1.ContainerStatus, error) {
	for _, status := range pod.Status.InitContainerStatuses {
		if status.State.Running != nil {
			return &status, nil
		}
		continue
	}
	return nil, fmt.Errorf("none of initContainers is running")
}

func getContainerStatus(pod *v1.Pod, cfg *config.Config) (*v1.ContainerStatus, error) {
	container := cfg.Container

	containerStatuses := make([]v1.ContainerStatus, 0)
	containerStatuses = append(containerStatuses, pod.Status.ContainerStatuses...)
	containerStatuses = append(containerStatuses, pod.Status.EphemeralContainerStatuses...)

	if container != "" {
		for _, status := range containerStatuses {
			if status.Name == container {
				if status.State.Running != nil {
					return &status, nil
				}
				return nil, fmt.Errorf("specified container is not running")
			}
			continue
		}
		return nil, fmt.Errorf("can't find specified container")

	}
	for _, status := range containerStatuses {
		if status.State.Running != nil {
			return &status, nil
		}
		continue
	}

	return nil, fmt.Errorf("pod %v has no running containers", pod.Name)
}

const containerdShellCmd = `"if command -v nerdctl >/dev/null 2>&1; then
	exec nerdctl inspect %[1]s --format {{.State.Pid}}
fi
exec crictl inspect --output go-template --template={{.info.pid}} %[1]s"`

func GetPidDiscoverCommand(containerInfo *ContainerInfo) ([]string, error) {
	switch containerInfo.ContainerRuntime {
	case "docker":
		return []string{
			"sudo",
			"docker",
			"inspect",
			containerInfo.ContainerID,
			"--format",
			"{{.State.Pid}}",
		}, nil

	case "containerd":
		return []string{"sudo", "sh", "-c", fmt.Sprintf(containerdShellCmd, containerInfo.ContainerID)}, nil

	case "cri-o":
		return []string{
			"sudo",
			"crictl",
			"inspect",
			"--output",
			"go-template",
			"--template={{.info.pid}}",
			containerInfo.ContainerID,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported container runtime")
	}
}
