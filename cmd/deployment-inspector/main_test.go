package main

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestParseTolerations(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []corev1.Toleration
		expectError bool
	}{
		{
			name:  "simple key=value:effect format",
			input: "role=myrole:NoSchedule",
			expected: []corev1.Toleration{
				{
					Key:      "role",
					Operator: corev1.TolerationOpEqual,
					Value:    "myrole",
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
			expectError: false,
		},
		{
			name:  "simple key:effect format (no value)",
			input: "role:NoSchedule",
			expected: []corev1.Toleration{
				{
					Key:      "role",
					Operator: corev1.TolerationOpExists,
					Value:    "",
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
			expectError: false,
		},
		{
			name:  "multiple tolerations",
			input: "role=myrole:NoSchedule,env=test:PreferNoSchedule",
			expected: []corev1.Toleration{
				{
					Key:      "role",
					Operator: corev1.TolerationOpEqual,
					Value:    "myrole",
					Effect:   corev1.TaintEffectNoSchedule,
				},
				{
					Key:      "env",
					Operator: corev1.TolerationOpEqual,
					Value:    "test",
					Effect:   corev1.TaintEffectPreferNoSchedule,
				},
			},
			expectError: false,
		},
		{
			name:  "case insensitive effects",
			input: "role=myrole:noschedule",
			expected: []corev1.Toleration{
				{
					Key:      "role",
					Operator: corev1.TolerationOpEqual,
					Value:    "myrole",
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
			expectError: false,
		},
		{
			name:  "NoExecute effect",
			input: "role=myrole:NoExecute",
			expected: []corev1.Toleration{
				{
					Key:      "role",
					Operator: corev1.TolerationOpEqual,
					Value:    "myrole",
					Effect:   corev1.TaintEffectNoExecute,
				},
			},
			expectError: false,
		},
		{
			name:  "JSON format",
			input: `[{"key":"role","operator":"Equal","value":"myrole","effect":"NoSchedule"}]`,
			expected: []corev1.Toleration{
				{
					Key:      "role",
					Operator: corev1.TolerationOpEqual,
					Value:    "myrole",
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
			expectError: false,
		},
		{
			name:        "invalid format - no colon",
			input:       "role=myrole",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid effect",
			input:       "role=myrole:InvalidEffect",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid JSON",
			input:       `[{"invalid": json}]`,
			expected:    nil,
			expectError: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: []corev1.Toleration{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTolerations(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tolerations, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				actual := result[i]
				if actual.Key != expected.Key {
					t.Errorf("Expected key %s, got %s", expected.Key, actual.Key)
				}
				if actual.Operator != expected.Operator {
					t.Errorf("Expected operator %s, got %s", expected.Operator, actual.Operator)
				}
				if actual.Value != expected.Value {
					t.Errorf("Expected value %s, got %s", expected.Value, actual.Value)
				}
				if actual.Effect != expected.Effect {
					t.Errorf("Expected effect %s, got %s", expected.Effect, actual.Effect)
				}
			}
		})
	}
}