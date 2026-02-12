package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stackgen-cli/devcheck/internal/models"
)

func TestDetectComposeFiles(t *testing.T) {
	// Create temp dir with compose file
	tmpDir, err := os.MkdirTemp("", "devcheck-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create compose.yaml
	if err := os.WriteFile(filepath.Join(tmpDir, "compose.yaml"), []byte("services: {}"), 0644); err != nil {
		t.Fatalf("failed to create compose.yaml: %v", err)
	}

	artifacts := Detect(tmpDir, "", nil)

	// Should find compose.yaml
	found := false
	for _, cf := range artifacts.ComposeFiles {
		if cf.Path == "compose.yaml" && cf.Found {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected to find compose.yaml")
	}
}

func TestDetectEnvFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "devcheck-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .env and .env.example
	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("KEY=value"), 0644); err != nil {
		t.Fatalf("failed to create .env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("KEY="), 0644); err != nil {
		t.Fatalf("failed to create .env.example: %v", err)
	}

	artifacts := Detect(tmpDir, "", nil)

	if !artifacts.HasEnv() {
		t.Error("expected to find .env")
	}
	if !artifacts.HasEnvExample() {
		t.Error("expected to find .env.example")
	}
}

func TestDetectLanguageManifests(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		content      string
		expectedLang models.Language
		expectedPM   string
	}{
		{
			name:         "Node.js with pnpm",
			file:         "package.json",
			content:      `{"name": "test"}`,
			expectedLang: models.LangNodeJS,
			expectedPM:   "",
		},
		{
			name:         "Go project",
			file:         "go.mod",
			content:      `module test`,
			expectedLang: models.LangGo,
			expectedPM:   "go mod",
		},
		{
			name:         "Python with pip",
			file:         "requirements.txt",
			content:      `flask>=2.0`,
			expectedLang: models.LangPython,
			expectedPM:   "pip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "devcheck-test")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			if err := os.WriteFile(filepath.Join(tmpDir, tt.file), []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create %s: %v", tt.file, err)
			}

			artifacts := Detect(tmpDir, "", nil)

			if artifacts.DetectedLang != tt.expectedLang {
				t.Errorf("expected language %s, got %s", tt.expectedLang, artifacts.DetectedLang)
			}
		})
	}
}

func TestDetectWithCustomCompose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "devcheck-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create custom compose file
	customFile := "my-compose.yaml"
	if err := os.WriteFile(filepath.Join(tmpDir, customFile), []byte("services: {}"), 0644); err != nil {
		t.Fatalf("failed to create %s: %v", customFile, err)
	}

	artifacts := Detect(tmpDir, customFile, nil)

	found := false
	for _, cf := range artifacts.ComposeFiles {
		if cf.Path == customFile && cf.Found {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected to find custom compose file %s", customFile)
	}
}

func TestDetectWithCustomEnvFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "devcheck-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create custom env files
	envFiles := []string{".env.custom", ".env.prod"}
	for _, ef := range envFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, ef), []byte("KEY=value"), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", ef, err)
		}
	}

	artifacts := Detect(tmpDir, "", envFiles)

	for _, ef := range envFiles {
		found := false
		for _, af := range artifacts.EnvFiles {
			if af.Path == ef && af.Found {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find custom env file %s", ef)
		}
	}
}
