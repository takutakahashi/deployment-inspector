package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DeploymentManagerInterface defines operations for deployment management
type DeploymentManagerInterface interface {
	GetPodsFromDeployment(deploymentName, namespace string) ([]corev1.Pod, error)
	GetNodesFromPods(pods []corev1.Pod) []string
}

// DeploymentManager manages deployment-related operations
type DeploymentManager struct {
	clientset *kubernetes.Clientset
}

// NewDeploymentManager creates a new deployment manager
func NewDeploymentManager(clientset *kubernetes.Clientset) DeploymentManagerInterface {
	return &DeploymentManager{
		clientset: clientset,
	}
}

// GetPodsFromDeployment returns all pods created by a specific deployment
func (dm *DeploymentManager) GetPodsFromDeployment(deploymentName, namespace string) ([]corev1.Pod, error) {
	deployment, err := dm.clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s: %v", deploymentName, err)
	}

	labelSelector := metav1.LabelSelector{MatchLabels: deployment.Spec.Selector.MatchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&labelSelector),
	}

	pods, err := dm.clientset.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	return pods.Items, nil
}

// GetNodesFromPods returns unique nodes where pods are running
func (dm *DeploymentManager) GetNodesFromPods(pods []corev1.Pod) []string {
	nodeSet := make(map[string]bool)
	for _, pod := range pods {
		if pod.Spec.NodeName != "" {
			nodeSet[pod.Spec.NodeName] = true
		}
	}

	nodes := make([]string, 0, len(nodeSet))
	for node := range nodeSet {
		nodes = append(nodes, node)
	}
	return nodes
}