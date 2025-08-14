package k8s

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestJobManager_CreateJobOnNodes(t *testing.T) {
	tests := []struct {
		name      string
		jobName   string
		nodes     []string
		namespace string
		image     string
		command   []string
		wantErr   bool
	}{
		{
			name:      "create job with default command",
			jobName:   "test-job",
			nodes:     []string{"node1", "node2"},
			namespace: "default",
			image:     "busybox",
			command:   nil,
			wantErr:   false,
		},
		{
			name:      "create job with custom command",
			jobName:   "test-job",
			nodes:     []string{"node1"},
			namespace: "custom-namespace",
			image:     "alpine",
			command:   []string{"ls", "-la"},
			wantErr:   false,
		},
		{
			name:      "create job with node containing dots",
			jobName:   "test-job",
			nodes:     []string{"node1.example.com"},
			namespace: "default",
			image:     "busybox",
			command:   []string{"echo", "test"},
			wantErr:   false,
		},
		{
			name:      "create job with empty nodes",
			jobName:   "test-job",
			nodes:     []string{},
			namespace: "default",
			image:     "busybox",
			command:   []string{"echo", "test"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake clientset
			clientset := fake.NewSimpleClientset()
			jm := &JobManager{clientset: clientset}

			// Execute
			jobs, err := jm.CreateJobOnNodes(tt.jobName, tt.nodes, tt.namespace, tt.image, tt.command)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateJobOnNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check number of jobs created
			if len(jobs) != len(tt.nodes) {
				t.Errorf("Expected %d jobs, got %d", len(tt.nodes), len(jobs))
			}

			// Verify jobs were created with correct configuration
			for i, node := range tt.nodes {
				jobName := jobs[i]
				
				// Get the created job
				job, err := clientset.BatchV1().Jobs(tt.namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
				if err != nil {
					t.Errorf("Failed to get created job %s: %v", jobName, err)
					continue
				}

				// Verify namespace
				if job.Namespace != tt.namespace {
					t.Errorf("Expected namespace %s, got %s", tt.namespace, job.Namespace)
				}

				// Verify TTL is set
				if job.Spec.TTLSecondsAfterFinished == nil {
					t.Error("TTLSecondsAfterFinished should be set")
				} else if *job.Spec.TTLSecondsAfterFinished != 300 {
					t.Errorf("Expected TTL 300, got %d", *job.Spec.TTLSecondsAfterFinished)
				}

				// Verify node selector
				if job.Spec.Template.Spec.NodeSelector["kubernetes.io/hostname"] != node {
					t.Errorf("Expected node selector %s, got %s", node, job.Spec.Template.Spec.NodeSelector["kubernetes.io/hostname"])
				}

				// Verify image
				if len(job.Spec.Template.Spec.Containers) > 0 {
					if job.Spec.Template.Spec.Containers[0].Image != tt.image {
						t.Errorf("Expected image %s, got %s", tt.image, job.Spec.Template.Spec.Containers[0].Image)
					}
				}

				// Verify command
				expectedCmd := tt.command
				if len(expectedCmd) == 0 {
					expectedCmd = []string{"echo", "Job running on node"}
				}
				
				if len(job.Spec.Template.Spec.Containers) > 0 {
					actualCmd := job.Spec.Template.Spec.Containers[0].Command
					if len(actualCmd) != len(expectedCmd) {
						t.Errorf("Expected command length %d, got %d", len(expectedCmd), len(actualCmd))
					}
				}

				// Verify job-name label matches job instance name
				if job.Spec.Template.Labels["job-name"] != jobName {
					t.Errorf("Expected job-name label %s, got %s", jobName, job.Spec.Template.Labels["job-name"])
				}
			}
		})
	}
}

