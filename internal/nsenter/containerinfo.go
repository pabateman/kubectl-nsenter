package nsenter

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type ContainerInfo struct {
	NodeName         string
	NodeIP           string
	ContainerID      string
	ContainerRuntime string
}

func NewPodSpec(name, namespace string, kubeClient *kubernetes.Clientset) (*v1.Pod, error) {
	return kubeClient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func PodInitialized(pod *v1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodInitialized {
			if condition.Status == v1.ConditionTrue {
				return true
			} else {
				return false
			}
		}
	}
	return false
}

func GetInitializingContainerStatus(pod *v1.Pod) (*v1.ContainerStatus, error) {
	for _, status := range pod.Status.InitContainerStatuses {
		if status.State.Running != nil {
			return &status, nil
		} else {
			continue
		}
	}
	return nil, fmt.Errorf("none of initContainers is running")
}

func GetContainerStatus(pod *v1.Pod, clictx *cli.Context) (*v1.ContainerStatus, error) {
	container := clictx.String("container")

	containerStatuses := make([]v1.ContainerStatus, 0)
	containerStatuses = append(containerStatuses, pod.Status.ContainerStatuses...)
	containerStatuses = append(containerStatuses, pod.Status.EphemeralContainerStatuses...)

	if container != "" {
		for _, status := range containerStatuses {
			if status.Name == container {
				if status.State.Running != nil {
					return &status, nil
				} else {
					return nil, fmt.Errorf("specified container is not running")
				}
			}
			continue
		}
		return nil, fmt.Errorf("can't find specified container")

	} else {
		for _, status := range containerStatuses {
			if status.State.Running != nil {
				return &status, nil
			}
			continue
		}
	}

	return nil, fmt.Errorf("pod %v has no running containers", pod.Name)
}

func GetClientSet(clictx *cli.Context) (*kubernetes.Clientset, string, error) {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{Precedence: strings.Split(clictx.String("kubeconfig"), ":")},
		&clientcmd.ConfigOverrides{
			CurrentContext: clictx.String("context"),
		})

	clientConfig, err := config.ClientConfig()
	if err != nil {
		return nil, "", errors.Wrap(err, "can't build config")
	}

	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, "", errors.Wrap(err, "can't build client")
	}

	namespace, _, err := config.Namespace()
	if err != nil {
		return nil, "", errors.Wrap(err, "can't get current namespace")
	}

	return clientset, namespace, nil
}

func GetContainerInfo(clictx *cli.Context) (*ContainerInfo, error) {
	kubeClient, namespace, err := GetClientSet(clictx)
	if err != nil {
		return nil, errors.Wrap(err, "can't build client")
	}

	if clictx.String("namespace") != "" {
		namespace = clictx.String("namespace")
	}

	podSpec, err := NewPodSpec(clictx.Args().First(), namespace, kubeClient)
	if err != nil {
		return nil, errors.Wrap(err, "can't get pod spec")
	}

	var containerStatus *v1.ContainerStatus

	if !PodInitialized(podSpec) {
		containerStatus, err = GetInitializingContainerStatus(podSpec)
		if err != nil {
			return nil, fmt.Errorf("pod is initializing: %v", err)
		}
	} else {
		containerStatus, err = GetContainerStatus(podSpec, clictx)
		if err != nil {
			return nil, fmt.Errorf("can't get container status: %v", err)
		}
	}

	return &ContainerInfo{
		NodeName:         podSpec.Spec.NodeName,
		NodeIP:           podSpec.Status.HostIP,
		ContainerID:      path.Base(containerStatus.ContainerID),
		ContainerRuntime: strings.Split(containerStatus.ContainerID, ":")[0],
	}, nil
}
