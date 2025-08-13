package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func getKubeClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		var kubeconfig string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
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

func getPodsFromDeployment(clientset *kubernetes.Clientset, deploymentName, namespace string) ([]corev1.Pod, error) {
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s: %v", deploymentName, err)
	}

	labelSelector := metav1.LabelSelector{MatchLabels: deployment.Spec.Selector.MatchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&labelSelector),
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	return pods.Items, nil
}

func getNodesFromPods(pods []corev1.Pod) []string {
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

func createJobOnNodes(clientset *kubernetes.Clientset, jobName string, nodes []string, namespace, image string, command []string) ([]string, error) {
	if len(command) == 0 {
		command = []string{"echo", "Job running on node"}
	}

	var jobsCreated []string

	for i, node := range nodes {
		jobInstanceName := fmt.Sprintf("%s-%s-%d", jobName, strings.ReplaceAll(node, ".", "-"), i)
		
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobInstanceName,
				Namespace: namespace,
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"job-name": jobName,
						},
					},
					Spec: corev1.PodSpec{
						RestartPolicy: corev1.RestartPolicyNever,
						NodeSelector: map[string]string{
							"kubernetes.io/hostname": node,
						},
						Containers: []corev1.Container{
							{
								Name:    "job-container",
								Image:   image,
								Command: command,
							},
						},
					},
				},
			},
		}

		_, err := clientset.BatchV1().Jobs(namespace).Create(context.TODO(), job, metav1.CreateOptions{})
		if err != nil {
			log.Printf("Error creating job on node %s: %v", node, err)
			continue
		}

		jobsCreated = append(jobsCreated, jobInstanceName)
		fmt.Printf("Created job %s on node %s\n", jobInstanceName, node)
	}

	return jobsCreated, nil
}

func listPodsAndNodes(clientset *kubernetes.Clientset, deploymentName, namespace string) ([]corev1.Pod, []string, error) {
	pods, err := getPodsFromDeployment(clientset, deploymentName, namespace)
	if err != nil {
		return nil, nil, err
	}

	if len(pods) == 0 {
		fmt.Printf("No pods found for deployment %s\n", deploymentName)
		return pods, nil, nil
	}

	fmt.Printf("\nPods from deployment '%s':\n", deploymentName)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%-40s %-20s\n", "Pod Name", "Node")
	fmt.Println(strings.Repeat("-", 60))

	for _, pod := range pods {
		node := pod.Spec.NodeName
		if node == "" {
			node = "Pending"
		}
		fmt.Printf("%-40s %-20s\n", pod.Name, node)
	}

	nodes := getNodesFromPods(pods)

	fmt.Printf("\nUnique nodes running pods from deployment '%s':\n", deploymentName)
	fmt.Println(strings.Repeat("-", 30))
	for _, node := range nodes {
		fmt.Printf("  - %s\n", node)
	}

	return pods, nodes, nil
}

func main() {
	var (
		namespace string
		image     string
		command   string
	)

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listCmd.StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	listCmd.StringVar(&namespace, "n", "default", "Kubernetes namespace (shorthand)")

	runJobCmd := flag.NewFlagSet("run-job", flag.ExitOnError)
	runJobCmd.StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	runJobCmd.StringVar(&namespace, "n", "default", "Kubernetes namespace (shorthand)")
	runJobCmd.StringVar(&image, "image", "busybox", "Container image for the job")
	runJobCmd.StringVar(&image, "i", "busybox", "Container image for the job (shorthand)")
	runJobCmd.StringVar(&command, "command", "", "Command to run in the job (comma-separated)")
	runJobCmd.StringVar(&command, "c", "", "Command to run in the job (comma-separated, shorthand)")

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  k8s-pod-node-job list <deployment-name> [-n namespace]")
		fmt.Println("  k8s-pod-node-job run-job <deployment-name> <job-name> [-n namespace] [-i image] [-c command]")
		os.Exit(1)
	}

	clientset, err := getKubeClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	switch os.Args[1] {
	case "list":
		listCmd.Parse(os.Args[2:])
		args := listCmd.Args()
		if len(args) < 1 {
			fmt.Println("Error: deployment name is required")
			os.Exit(1)
		}
		deploymentName := args[0]

		_, _, err := listPodsAndNodes(clientset, deploymentName, namespace)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

	case "run-job":
		runJobCmd.Parse(os.Args[2:])
		args := runJobCmd.Args()
		if len(args) < 2 {
			fmt.Println("Error: deployment name and job name are required")
			os.Exit(1)
		}
		deploymentName := args[0]
		jobName := args[1]

		pods, err := getPodsFromDeployment(clientset, deploymentName, namespace)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		if len(pods) == 0 {
			fmt.Printf("No pods found for deployment %s\n", deploymentName)
			os.Exit(1)
		}

		nodes := getNodesFromPods(pods)
		if len(nodes) == 0 {
			fmt.Println("No nodes found with running pods")
			os.Exit(1)
		}

		fmt.Printf("\nCreating jobs on %d nodes...\n", len(nodes))
		
		var cmdSlice []string
		if command != "" {
			cmdSlice = strings.Split(command, ",")
			for i := range cmdSlice {
				cmdSlice[i] = strings.TrimSpace(cmdSlice[i])
			}
		}

		jobs, err := createJobOnNodes(clientset, jobName, nodes, namespace, image, cmdSlice)
		if err != nil {
			log.Printf("Warning: %v", err)
		}

		if len(jobs) > 0 {
			fmt.Printf("\nSuccessfully created %d jobs\n", len(jobs))
		} else {
			fmt.Println("\nNo jobs were created")
		}

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Println("Use 'list' or 'run-job'")
		os.Exit(1)
	}
}