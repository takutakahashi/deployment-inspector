package k8s

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeploymentManager_GetNodesFromPods(t *testing.T) {
	tests := []struct {
		name     string
		pods     []corev1.Pod
		expected []string
	}{
		{
			name:     "empty pods",
			pods:     []corev1.Pod{},
			expected: []string{},
		},
		{
			name: "pods without node names",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
					Spec:       corev1.PodSpec{NodeName: ""},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod2"},
					Spec:       corev1.PodSpec{NodeName: ""},
				},
			},
			expected: []string{},
		},
		{
			name: "pods with unique nodes",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
					Spec:       corev1.PodSpec{NodeName: "node1"},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod2"},
					Spec:       corev1.PodSpec{NodeName: "node2"},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod3"},
					Spec:       corev1.PodSpec{NodeName: "node3"},
				},
			},
			expected: []string{"node1", "node2", "node3"},
		},
		{
			name: "pods with duplicate nodes",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
					Spec:       corev1.PodSpec{NodeName: "node1"},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod2"},
					Spec:       corev1.PodSpec{NodeName: "node1"},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod3"},
					Spec:       corev1.PodSpec{NodeName: "node2"},
				},
			},
			expected: []string{"node1", "node2"},
		},
		{
			name: "mixed pods with and without nodes",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
					Spec:       corev1.PodSpec{NodeName: "node1"},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod2"},
					Spec:       corev1.PodSpec{NodeName: ""},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod3"},
					Spec:       corev1.PodSpec{NodeName: "node2"},
				},
			},
			expected: []string{"node1", "node2"},
		},
	}

	dm := &DeploymentManager{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes := dm.GetNodesFromPods(tt.pods)
			
			if len(nodes) != len(tt.expected) {
				t.Errorf("GetNodesFromPods() returned %d nodes, expected %d", len(nodes), len(tt.expected))
				return
			}
			
			// Create a map for easy lookup
			nodeMap := make(map[string]bool)
			for _, node := range nodes {
				nodeMap[node] = true
			}
			
			// Check all expected nodes are present
			for _, expectedNode := range tt.expected {
				if !nodeMap[expectedNode] {
					t.Errorf("Expected node %s not found in result", expectedNode)
				}
			}
		})
	}
}