package services

import (
	"testing"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected Status
	}{
		{"active", StatusActive},
		{"inactive", StatusInactive},
		{"failed", StatusFailed},
		{"activating", StatusActivating},
		{"deactivating", StatusDeactivating},
		{"unknown", StatusUnknown},
		{"", StatusUnknown},
		{"random", StatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseStatus(tt.input)
			if result != tt.expected {
				t.Errorf("ParseStatus(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestServiceIsRunning(t *testing.T) {
	svc := Service{Name: "nginx", Status: StatusActive}
	if !svc.IsRunning() {
		t.Error("Service active devrait être IsRunning()")
	}

	svc.Status = StatusInactive
	if svc.IsRunning() {
		t.Error("Service inactive ne devrait pas être IsRunning()")
	}
}

func TestServiceIsFailed(t *testing.T) {
	svc := Service{Name: "myapp", Status: StatusFailed}
	if !svc.IsFailed() {
		t.Error("Service failed devrait être IsFailed()")
	}

	svc.Status = StatusActive
	if svc.IsFailed() {
		t.Error("Service active ne devrait pas être IsFailed()")
	}
}

func TestServiceStatusLabel(t *testing.T) {
	tests := []struct {
		status Status
		label  string
	}{
		{StatusActive, "● Actif"},
		{StatusInactive, "○ Inactif"},
		{StatusFailed, "✖ Erreur"},
		{StatusActivating, "▶ Démarrage..."},
		{StatusDeactivating, "■ Arrêt..."},
		{StatusUnknown, "? Inconnu"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			svc := Service{Status: tt.status}
			if got := svc.StatusLabel(); got != tt.label {
				t.Errorf("StatusLabel() = %q, want %q", got, tt.label)
			}
		})
	}
}
