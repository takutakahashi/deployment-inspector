package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/takutakahashi/deployment-inspector/pkg/k8s"
)

func listPodsAndNodes(deploymentName, namespace string) error {
	client := k8s.NewClient("")
	clientset, err := client.GetClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	deploymentManager := k8s.NewDeploymentManager(clientset)
	
	pods, err := deploymentManager.GetPodsFromDeployment(deploymentName, namespace)
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		fmt.Printf("No pods found for deployment %s\n", deploymentName)
		return nil
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

	nodes := deploymentManager.GetNodesFromPods(pods)

	fmt.Printf("\nUnique nodes running pods from deployment '%s':\n", deploymentName)
	fmt.Println(strings.Repeat("-", 30))
	for _, node := range nodes {
		fmt.Printf("  - %s\n", node)
	}

	return nil
}

func runJobOnNodes(deploymentName, jobName, namespace, jobNamespace, image string, command []string) error {
	client := k8s.NewClient("")
	clientset, err := client.GetClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	deploymentManager := k8s.NewDeploymentManager(clientset)
	jobManager := k8s.NewJobManager(clientset)
	
	pods, err := deploymentManager.GetPodsFromDeployment(deploymentName, namespace)
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		fmt.Printf("No pods found for deployment %s\n", deploymentName)
		return nil
	}

	nodes := deploymentManager.GetNodesFromPods(pods)
	if len(nodes) == 0 {
		fmt.Println("No nodes found with running pods")
		return nil
	}

	fmt.Printf("\nCreating jobs on %d nodes...\n", len(nodes))
	
	jobs, err := jobManager.CreateJobOnNodes(jobName, nodes, jobNamespace, image, command)
	if err != nil {
		log.Printf("Warning: %v", err)
	}

	for _, job := range jobs {
		fmt.Printf("Created job %s\n", job)
	}

	if len(jobs) > 0 {
		fmt.Printf("\nSuccessfully created %d jobs\n", len(jobs))
	} else {
		fmt.Println("\nNo jobs were created")
	}

	return nil
}

func main() {
	var (
		namespace    string
		jobNamespace string
		image        string
		command      string
	)

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listCmd.StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	listCmd.StringVar(&namespace, "n", "default", "Kubernetes namespace (shorthand)")

	runJobCmd := flag.NewFlagSet("run-job", flag.ExitOnError)
	runJobCmd.StringVar(&namespace, "namespace", "default", "Kubernetes namespace for deployment")
	runJobCmd.StringVar(&namespace, "n", "default", "Kubernetes namespace for deployment (shorthand)")
	runJobCmd.StringVar(&jobNamespace, "job-namespace", "", "Kubernetes namespace for job (defaults to deployment namespace)")
	runJobCmd.StringVar(&jobNamespace, "jn", "", "Kubernetes namespace for job (shorthand)")
	runJobCmd.StringVar(&image, "image", "busybox", "Container image for the job")
	runJobCmd.StringVar(&image, "i", "busybox", "Container image for the job (shorthand)")
	runJobCmd.StringVar(&command, "command", "", "Command to run in the job (comma-separated)")
	runJobCmd.StringVar(&command, "c", "", "Command to run in the job (comma-separated, shorthand)")

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  deployment-inspector list <deployment-name> [-n namespace]")
		fmt.Println("  deployment-inspector run-job <deployment-name> <job-name> [-n namespace] [-jn job-namespace] [-i image] [-c command]")
		os.Exit(1)
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

		if err := listPodsAndNodes(deploymentName, namespace); err != nil {
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

		// If job namespace is not specified, use the deployment namespace
		if jobNamespace == "" {
			jobNamespace = namespace
		}

		var cmdSlice []string
		if command != "" {
			cmdSlice = strings.Split(command, ",")
			for i := range cmdSlice {
				cmdSlice[i] = strings.TrimSpace(cmdSlice[i])
			}
		}

		if err := runJobOnNodes(deploymentName, jobName, namespace, jobNamespace, image, cmdSlice); err != nil {
			log.Fatalf("Error: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Println("Use 'list' or 'run-job'")
		os.Exit(1)
	}
}