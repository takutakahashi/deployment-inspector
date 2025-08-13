package k8s

import (
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ClientInterface defines the interface for Kubernetes client operations
type ClientInterface interface {
	GetClient() (*kubernetes.Clientset, error)
}

// Client implements the ClientInterface
type Client struct {
	kubeconfig string
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfig string) ClientInterface {
	return &Client{
		kubeconfig: kubeconfig,
	}
}

// GetClient returns a configured Kubernetes clientset
func (c *Client) GetClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := c.kubeconfig
		if kubeconfig == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			}
		}
		
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}