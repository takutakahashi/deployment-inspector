package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/takutakahashi/deployment-inspector/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
)

var (
	rootCmd = &cobra.Command{
		Use:   "deployment-inspector",
		Short: "A tool to inspect Kubernetes deployments and run jobs on their nodes",
		Long: `deployment-inspector is a CLI tool that helps you inspect Kubernetes deployments
and run jobs on the nodes where deployment pods are running.`,
	}

	listCmd = &cobra.Command{
		Use:   "list <deployment-name>",
		Short: "List pods and nodes for a deployment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deploymentName := args[0]
			namespace := viper.GetString("namespace")
			return listPodsAndNodes(deploymentName, namespace)
		},
	}

	runJobCmd = &cobra.Command{
		Use:   "run-job <deployment-name> <job-name>",
		Short: "Run a job on nodes where deployment pods are running",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			deploymentName := args[0]
			jobName := args[1]
			namespace := viper.GetString("namespace")
			jobNamespace := viper.GetString("job-namespace")
			image := viper.GetString("image")
			commandStr := viper.GetString("command")
			tolerationsStr := viper.GetString("tolerations")

			// If job namespace is not specified, use the deployment namespace
			if jobNamespace == "" {
				jobNamespace = namespace
			}

			var command []string
			if commandStr != "" {
				command = strings.Split(commandStr, ",")
				for i := range command {
					command[i] = strings.TrimSpace(command[i])
				}
			}

			// Parse tolerations from JSON string
			var tolerations []corev1.Toleration
			if tolerationsStr != "" {
				err := json.Unmarshal([]byte(tolerationsStr), &tolerations)
				if err != nil {
					return fmt.Errorf("failed to parse tolerations JSON: %v", err)
				}
			}

			return runJobOnNodes(deploymentName, jobName, namespace, jobNamespace, image, command, tolerations)
		},
	}
)

func init() {
	// Persistent flags available to all commands
	rootCmd.PersistentFlags().StringP("namespace", "n", "default", "Kubernetes namespace")
	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))

	// Run-job specific flags
	runJobCmd.Flags().StringP("job-namespace", "j", "", "Kubernetes namespace for job (defaults to deployment namespace)")
	runJobCmd.Flags().StringP("image", "i", "busybox", "Container image for the job")
	runJobCmd.Flags().StringP("command", "c", "", "Command to run in the job (comma-separated)")
	runJobCmd.Flags().StringP("tolerations", "t", "", "Tolerations for the job pods (JSON format)")
	
	viper.BindPFlag("job-namespace", runJobCmd.Flags().Lookup("job-namespace"))
	viper.BindPFlag("image", runJobCmd.Flags().Lookup("image"))
	viper.BindPFlag("command", runJobCmd.Flags().Lookup("command"))
	viper.BindPFlag("tolerations", runJobCmd.Flags().Lookup("tolerations"))

	// Add commands to root
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(runJobCmd)
}

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
		fmt.Printf("No pods found for deployment %s in namespace %s\n", deploymentName, namespace)
		return nil
	}

	fmt.Printf("\nPods from deployment '%s' in namespace '%s':\n", deploymentName, namespace)
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

func runJobOnNodes(deploymentName, jobName, namespace, jobNamespace, image string, command []string, tolerations []corev1.Toleration) error {
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
		fmt.Printf("No pods found for deployment %s in namespace %s\n", deploymentName, namespace)
		return nil
	}

	nodes := deploymentManager.GetNodesFromPods(pods)
	if len(nodes) == 0 {
		fmt.Println("No nodes found with running pods")
		return nil
	}

	fmt.Printf("\nCreating jobs on %d nodes in namespace %s...\n", len(nodes), jobNamespace)
	
	jobs, err := jobManager.CreateJobOnNodes(jobName, nodes, jobNamespace, image, command, tolerations)
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
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}