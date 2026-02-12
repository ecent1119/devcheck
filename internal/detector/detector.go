package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/stackgen-cli/devcheck/internal/models"
)

// Detect scans a directory for project artifacts
func Detect(basePath string, composeOverride string, envOverrides []string) *models.Artifacts {
	artifacts := models.NewArtifacts()

	// Detect compose files
	detectComposeFiles(basePath, composeOverride, artifacts)

	// Detect env files
	detectEnvFiles(basePath, envOverrides, artifacts)

	// Detect language manifests
	detectManifests(basePath, artifacts)

	// Detect README
	detectReadme(basePath, artifacts)

	// Detect Makefile
	detectMakefile(basePath, artifacts)

	return artifacts
}

// detectComposeFiles looks for Docker Compose files
func detectComposeFiles(basePath string, override string, artifacts *models.Artifacts) {
	// Check override first
	if override != "" {
		fullPath := override
		if !filepath.IsAbs(override) {
			fullPath = filepath.Join(basePath, override)
		}
		found := fileExists(fullPath)
		artifacts.ComposeFiles = append(artifacts.ComposeFiles, models.Artifact{
			Type:  models.ArtifactCompose,
			Path:  override,
			Found: found,
		})
		if found {
			return // Only use override if specified
		}
	}

	// Standard compose file names
	candidates := []string{
		"compose.yaml",
		"compose.yml",
		"docker-compose.yaml",
		"docker-compose.yml",
		"docker-compose.override.yaml",
		"docker-compose.override.yml",
	}

	for _, name := range candidates {
		fullPath := filepath.Join(basePath, name)
		found := fileExists(fullPath)
		if found {
			artifacts.ComposeFiles = append(artifacts.ComposeFiles, models.Artifact{
				Type:  models.ArtifactCompose,
				Path:  name,
				Found: true,
			})
		}
	}
}

// detectEnvFiles looks for environment files
func detectEnvFiles(basePath string, overrides []string, artifacts *models.Artifacts) {
	// Check overrides first
	if len(overrides) > 0 {
		for _, override := range overrides {
			fullPath := override
			if !filepath.IsAbs(override) {
				fullPath = filepath.Join(basePath, override)
			}
			found := fileExists(fullPath)
			if strings.Contains(override, "example") {
				artifacts.EnvExamples = append(artifacts.EnvExamples, models.Artifact{
					Type:  models.ArtifactEnvExample,
					Path:  override,
					Found: found,
				})
			} else {
				artifacts.EnvFiles = append(artifacts.EnvFiles, models.Artifact{
					Type:  models.ArtifactEnv,
					Path:  override,
					Found: found,
				})
			}
		}
		return
	}

	// Standard env file names
	envCandidates := []string{
		".env",
		".env.local",
		".env.development",
		".env.dev",
	}

	exampleCandidates := []string{
		".env.example",
		".env.sample",
		".env.template",
		"example.env",
	}

	for _, name := range envCandidates {
		fullPath := filepath.Join(basePath, name)
		found := fileExists(fullPath)
		artifacts.EnvFiles = append(artifacts.EnvFiles, models.Artifact{
			Type:  models.ArtifactEnv,
			Path:  name,
			Found: found,
		})
	}

	for _, name := range exampleCandidates {
		fullPath := filepath.Join(basePath, name)
		found := fileExists(fullPath)
		if found {
			artifacts.EnvExamples = append(artifacts.EnvExamples, models.Artifact{
				Type:  models.ArtifactEnvExample,
				Path:  name,
				Found: true,
			})
		}
	}
}

// detectManifests looks for language-specific manifest files
func detectManifests(basePath string, artifacts *models.Artifacts) {
	manifests := []struct {
		file    string
		lang    models.Language
		pkgMgr  string
		details string
	}{
		// Node.js
		{"package.json", models.LangNodeJS, "", "Node.js project"},
		{"pnpm-lock.yaml", models.LangNodeJS, "pnpm", "pnpm lockfile"},
		{"yarn.lock", models.LangNodeJS, "yarn", "Yarn lockfile"},
		{"package-lock.json", models.LangNodeJS, "npm", "npm lockfile"},

		// Go
		{"go.mod", models.LangGo, "go mod", "Go module"},

		// Python
		{"pyproject.toml", models.LangPython, "", "Python project"},
		{"requirements.txt", models.LangPython, "pip", "pip requirements"},
		{"Pipfile", models.LangPython, "pipenv", "Pipenv project"},
		{"poetry.lock", models.LangPython, "poetry", "Poetry project"},

		// Rust
		{"Cargo.toml", models.LangRust, "cargo", "Rust project"},

		// Java
		{"pom.xml", models.LangJava, "maven", "Maven project"},
		{"build.gradle", models.LangJava, "gradle", "Gradle project"},
		{"build.gradle.kts", models.LangJava, "gradle", "Gradle Kotlin project"},

		// C#
		{"*.csproj", models.LangCSharp, "dotnet", "C# project"},
		{"*.sln", models.LangCSharp, "dotnet", "C# solution"},
	}

	for _, m := range manifests {
		var found bool
		var actualPath string

		if strings.Contains(m.file, "*") {
			// Glob pattern
			matches, _ := filepath.Glob(filepath.Join(basePath, m.file))
			found = len(matches) > 0
			if found {
				actualPath = filepath.Base(matches[0])
			}
		} else {
			fullPath := filepath.Join(basePath, m.file)
			found = fileExists(fullPath)
			actualPath = m.file
		}

		if found {
			artifacts.Manifests = append(artifacts.Manifests, models.Artifact{
				Type:     models.ArtifactManifest,
				Path:     actualPath,
				Language: m.lang,
				Details:  m.details,
				Found:    true,
			})

			// Set primary language (first found wins)
			if artifacts.DetectedLang == "" {
				artifacts.DetectedLang = m.lang
			}

			// Set package manager if more specific
			if m.pkgMgr != "" && artifacts.PackageManager == "" {
				artifacts.PackageManager = m.pkgMgr
			}
		}
	}
}

// detectReadme looks for README files
func detectReadme(basePath string, artifacts *models.Artifacts) {
	candidates := []string{
		"README.md",
		"README.MD",
		"readme.md",
		"README.txt",
		"README",
	}

	for _, name := range candidates {
		fullPath := filepath.Join(basePath, name)
		if fileExists(fullPath) {
			artifacts.Readme = &models.Artifact{
				Type:  models.ArtifactReadme,
				Path:  name,
				Found: true,
			}
			return
		}
	}
}

// detectMakefile looks for Makefile
func detectMakefile(basePath string, artifacts *models.Artifacts) {
	candidates := []string{
		"Makefile",
		"makefile",
		"GNUmakefile",
	}

	for _, name := range candidates {
		fullPath := filepath.Join(basePath, name)
		if fileExists(fullPath) {
			artifacts.Makefile = &models.Artifact{
				Type:  models.ArtifactMakefile,
				Path:  name,
				Found: true,
			}
			return
		}
	}
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
