package kubernetes

// TODO : DOCU

import (
	"context"
	"fmt"
	"sync"

	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	once           sync.Once
	clientInstance *kubernetes.Clientset
)

// initK8sClient initializes the Kubernetes client.
// It first attempts to create an in-cluster configuration. If that fails (indicating that the code is not running inside a Kubernetes cluster),
// it falls back to using a kubeconfig file specified by an environment variable.
func initK8sClient() (*kubernetes.Clientset, error) {
	// Create the Kubernetes client configuration
	k8sClientConfig, err := rest.InClusterConfig()

	// TODO - Config value for the kubeconfig path
	// TODO - Key off config instead of dynaminc fall back
	// If not running inside a Kubernetes cluster, fall back to using the kubeconfig file
	if err != nil {
		kubeconfigPath := config.GetKubeConfigPath()
		k8sClientConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to load Kubernetes configuration. Root Cause: %v", err)
		}
	}

	k8sClient, err := kubernetes.NewForConfig(k8sClientConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Kubernetes client. Root Cause: %v", err)
	}

	return k8sClient, nil
}

// getK8sClient returns the initialized Kubernetes client. It uses a sync.Once to ensure that the client is only initialized once.
func getK8sClient() (*kubernetes.Clientset, error) {
	var err error
	once.Do(func() {
		clientInstance, err = initK8sClient()
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to initialize Kubernetes client. Root Cause: %v", err)
	}

	return clientInstance, nil
}

func GetPodLogs(namespace string, podName string, sinceTime *metav1.Time, lastLines *int64) (string, error) {
	k8sClient, err := getK8sClient()

	if err != nil {
		return "", fmt.Errorf("Failed to get Kubernetes client. Root Cause: %v", err)
	}

	logs, err := k8sClient.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{
		SinceTime: sinceTime,
		TailLines: lastLines,
	}).Do(context.Background()).Raw()

	if err != nil {
		return "", fmt.Errorf("Failed to get logs for pod %s in namespace %s. Root Cause: %v", podName, namespace, err)
	}

	return string(logs), nil
}
