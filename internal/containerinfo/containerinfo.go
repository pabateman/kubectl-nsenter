package containerinfo

import (
	"context"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type PodContext struct {
	Clientset *kubernetes.Clientset
	Namespace string
}

type ContainerInfo struct {
	NodeName         string
	NodeIP           string
	ContainerID      string
	ContainerRuntime string
}

func GetPodContext(kubeconfigFiles []string) (*kubernetes.Clientset, string, error) {
	// use the current context in kubeconfig
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{Precedence: kubeconfigFiles},
		&clientcmd.ConfigOverrides{})
	namespace, _, err := config.Namespace()
	if err != nil {
		return nil, "", errors.Wrap(err, "can't get current namespace")
	}
	clientConfig, err := config.ClientConfig()
	if err != nil {
		return nil, "", errors.Wrap(err, "can't build config")
	}
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, "", errors.Wrap(err, "can't build client")
	}
	return clientset, namespace, nil
}

func GetContainerInfo(k []string, p string, c string) (*ContainerInfo, error) {

	clientset, namespace, err := GetPodContext(k)
	podSpec, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), p, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "can't get pod spec")
	}

	var containerId string
	var containerRuntime string
	if c != "" {
		for _, containerStatus := range podSpec.Status.ContainerStatuses {
			if containerStatus.Name == c {
				containerId = path.Base(containerStatus.ContainerID)
				containerRuntime = strings.Split(containerStatus.ContainerID, ":")[0]
			}
		}
	} else {
		container := podSpec.Status.ContainerStatuses[0].ContainerID
		containerId = path.Base(container)
		containerRuntime = strings.Split(container, ":")[0]
	}

	return &ContainerInfo{
		NodeName:         podSpec.Spec.NodeName,
		NodeIP:           podSpec.Status.HostIP,
		ContainerID:      containerId,
		ContainerRuntime: containerRuntime,
	}, nil
}
