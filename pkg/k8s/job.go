package k8s

import (
	"context"
	"fmt"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// JobManagerInterface defines operations for job management
type JobManagerInterface interface {
	CreateJobOnNodes(jobName string, nodes []string, namespace, image string, command []string) ([]string, error)
}

// JobManager manages job-related operations
type JobManager struct {
	clientset *kubernetes.Clientset
}

// NewJobManager creates a new job manager
func NewJobManager(clientset *kubernetes.Clientset) JobManagerInterface {
	return &JobManager{
		clientset: clientset,
	}
}

// CreateJobOnNodes creates jobs on specified nodes
func (jm *JobManager) CreateJobOnNodes(jobName string, nodes []string, namespace, image string, command []string) ([]string, error) {
	if len(command) == 0 {
		command = []string{"echo", "Job running on node"}
	}

	var jobsCreated []string
	var lastError error

	for i, node := range nodes {
		jobInstanceName := fmt.Sprintf("%s-%s-%d", jobName, strings.ReplaceAll(node, ".", "-"), i)
		
		ttlSecondsAfterFinished := int32(300) // 5 minutes after completion
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobInstanceName,
				Namespace: namespace,
			},
			Spec: batchv1.JobSpec{
				TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"job-name": jobInstanceName,
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

		_, err := jm.clientset.BatchV1().Jobs(namespace).Create(context.TODO(), job, metav1.CreateOptions{})
		if err != nil {
			lastError = fmt.Errorf("failed to create job on node %s: %v", node, err)
			continue
		}

		jobsCreated = append(jobsCreated, jobInstanceName)
	}

	if len(jobsCreated) == 0 && lastError != nil {
		return nil, lastError
	}

	return jobsCreated, nil
}