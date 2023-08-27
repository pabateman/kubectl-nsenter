package k8s

import (
	"context"
	"strings"

	"github.com/pabateman/kubectl-nsenter/internal/config"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func GetKubernetesClient(cfg *config.Config) (*kubernetes.Clientset, error) {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{Precedence: strings.Split(cfg.KubeConfig, ":")},
		&clientcmd.ConfigOverrides{
			CurrentContext: cfg.KubeContext,
		})

	clientConfig, err := config.ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "can't build config")
	}

	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, errors.Wrap(err, "can't build client")
	}

	if cfg.Namespace == "" {
		cfg.Namespace, _, err = config.Namespace()
		if err != nil {
			return nil, errors.Wrap(err, "can't get current namespace")
		}
	}

	return clientset, nil
}

func GetPod(cfg *config.Config, kubeClient *kubernetes.Clientset) (*v1.Pod, error) {
	return kubeClient.CoreV1().Pods(cfg.Namespace).Get(context.TODO(), cfg.PodName, metav1.GetOptions{})
}
