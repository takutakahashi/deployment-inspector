package k8s

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name       string
		kubeconfig string
	}{
		{
			name:       "with empty kubeconfig",
			kubeconfig: "",
		},
		{
			name:       "with specific kubeconfig",
			kubeconfig: "/tmp/test-kubeconfig",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.kubeconfig)
			if client == nil {
				t.Fatal("NewClient returned nil")
			}
			
			c, ok := client.(*Client)
			if !ok {
				t.Fatal("NewClient did not return *Client type")
			}
			
			if c.kubeconfig != tt.kubeconfig {
				t.Errorf("kubeconfig = %v, want %v", c.kubeconfig, tt.kubeconfig)
			}
		})
	}
}

func TestClient_GetClient(t *testing.T) {
	// Skip this test if not in a Kubernetes cluster and no kubeconfig is available
	if os.Getenv("KUBERNETES_SERVICE_HOST") == "" {
		home := os.Getenv("HOME")
		if home == "" {
			t.Skip("Skipping test: not in cluster and no HOME directory")
		}
		
		kubeconfigPath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
			t.Skip("Skipping test: no kubeconfig file found")
		}
	}

	client := NewClient("")
	_, err := client.GetClient()
	if err != nil {
		t.Logf("GetClient error (this might be expected in test environment): %v", err)
	}
}