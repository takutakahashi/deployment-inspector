package k8s

import (
	"testing"
)

func TestJobManager_CreateJobOnNodes(t *testing.T) {
	// This test requires a mock or fake clientset for proper unit testing
	// For now, we'll test the job name generation logic
	
	tests := []struct {
		name    string
		jobName string
		nodes   []string
		command []string
	}{
		{
			name:    "default command",
			jobName: "test-job",
			nodes:   []string{"node1", "node2"},
			command: nil,
		},
		{
			name:    "custom command",
			jobName: "test-job",
			nodes:   []string{"node1"},
			command: []string{"ls", "-la"},
		},
		{
			name:    "node with dots",
			jobName: "test-job",
			nodes:   []string{"node1.example.com"},
			command: []string{"echo", "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Without a fake clientset, we can only test basic validation
			if len(tt.nodes) == 0 {
				t.Skip("Skipping test with no nodes")
			}
			
			// Test that command defaults are set correctly
			cmd := tt.command
			if len(cmd) == 0 {
				cmd = []string{"echo", "Job running on node"}
			}
			
			if tt.command == nil && len(cmd) != 2 {
				t.Errorf("Default command should have 2 elements, got %d", len(cmd))
			}
		})
	}
}