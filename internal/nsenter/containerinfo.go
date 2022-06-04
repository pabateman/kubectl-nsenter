package nsenter

import (
	"context"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
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

	podSpec, err := kubeClient.CoreV1().Pods(namespace).Get(context.TODO(), clictx.Args().First(), metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "can't get pod spec")
	}

	var containerID string
	var containerRuntime string
	container := clictx.String("container")

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
