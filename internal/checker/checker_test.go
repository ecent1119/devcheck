package checker

import (
	"path/filepath"
	"testing"

	"github.com/stackgen-cli/devcheck/internal/detector"
	"github.com/stackgen-cli/devcheck/internal/models"
)

func TestCheckBasicProject(t *testing.T) {
	basePath, err := filepath.Abs("testdata/basic")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	artifacts := detector.Detect(basePath, "", nil)
	findings := Check(basePath, artifacts)

	// Should have no blocking findings since all env vars are defined
	blocking := countByCode(findings, "ENV001")
	if blocking > 0 {
		t.Errorf("expected 0 ENV001 findings, got %d", blocking)
		for _, f := range findings {
			if f.Code == "ENV001" {
				t.Logf("  - %s: %s", f.Code, f.Title)
			}
		}
	}
}

func TestCheckMissingEnvVars(t *testing.T) {
	basePath, err := filepath.Abs("testdata/missing-env")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	artifacts := detector.Detect(basePath, "", nil)
	findings := Check(basePath, artifacts)

	// Should have 2 blocking findings for SECRET_TOKEN and REDIS_URL
	blocking := countByCode(findings, "ENV001")
	if blocking != 2 {
		t.Errorf("expected 2 ENV001 findings, got %d", blocking)
		for _, f := range findings {
			t.Logf("  - %s: %s", f.Code, f.Title)
		}
	}

	// Verify specific missing vars
	foundSecret := false
	foundRedis := false
	for _, f := range findings {
		if f.Code == "ENV001" {
			if contains(f.Title, "SECRET_TOKEN") {
				foundSecret = true
			}
			if contains(f.Title, "REDIS_URL") {
				foundRedis = true
			}
		}
	}

	if !foundSecret {
		t.Error("expected finding for SECRET_TOKEN not found")
	}
	if !foundRedis {
		t.Error("expected finding for REDIS_URL not found")
	}
}

func TestParseEnvFile(t *testing.T) {
	basePath, _ := filepath.Abs("testdata/basic")
	vars := parseEnvFile(filepath.Join(basePath, ".env"))

	if vars["DATABASE_HOST"] != "localhost" {
		t.Errorf("expected DATABASE_HOST=localhost, got %s", vars["DATABASE_HOST"])
	}
	if vars["API_KEY"] != "test-key" {
		t.Errorf("expected API_KEY=test-key, got %s", vars["API_KEY"])
	}
}

func TestIsStandardVar(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"HOME", true},
		{"PATH", true},
		{"CUSTOM_VAR", false},
		{"DATABASE_URL", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStandardVar(tt.name); got != tt.expected {
				t.Errorf("isStandardVar(%s) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

// Helper functions

func countByCode(findings []*models.Finding, code string) int {
	count := 0
	for _, f := range findings {
		if f.Code == code {
			count++
		}
	}
	return count
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
