package nsenter

import (
	"context"
	"path"
	"strings"

	"github.com/pkg/errors"
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

func GetContainerInfo(kubeconfigFiles []string, contextOverride string, namespaceOverride string, pod string, container string) (*ContainerInfo, error) {

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{Precedence: kubeconfigFiles},
		&clientcmd.ConfigOverrides{
			CurrentContext: contextOverride,
		})

	var namespace string
	var err error
	if namespaceOverride == "" {
		namespace, _, err = config.Namespace()
		if err != nil {
			return nil, errors.Wrap(err, "can't get current namespace")
		}
	} else {
		namespace = namespaceOverride
	}

	clientConfig, err := config.ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "can't build config")
	}
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, errors.Wrap(err, "can't build client")
	}

	podSpec, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "can't get pod spec")
	}

	var containerID string
	var containerRuntime string
	if container != "" {
		for _, containerStatus := range podSpec.Status.ContainerStatuses {
			if containerStatus.Name == container {
				containerID = path.Base(containerStatus.ContainerID)
				containerRuntime = strings.Split(containerStatus.ContainerID, ":")[0]
			}
		}
	} else {
		defaultContainer := podSpec.Status.ContainerStatuses[0].ContainerID
		containerID = path.Base(defaultContainer)
		containerRuntime = strings.Split(defaultContainer, ":")[0]
	}

	return &ContainerInfo{
		NodeName:         podSpec.Spec.NodeName,
		NodeIP:           podSpec.Status.HostIP,
		ContainerID:      containerID,
		ContainerRuntime: containerRuntime,
	}, nil
}
