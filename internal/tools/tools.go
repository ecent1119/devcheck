// Package tools handles tool version detection and validation
package tools

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// ToolInfo contains detected tool version information
type ToolInfo struct {
	Name      string
	Version   string
	Path      string
	Available bool
	Error     string
}

// VersionCheck result
type VersionCheck struct {
	Tool       string
	Current    string
	Required   string
	Satisfied  bool
	Available  bool
	Error      string
}

// DetectTools checks for common development tools
func DetectTools() map[string]ToolInfo {
	tools := make(map[string]ToolInfo)

	// Docker
	tools["docker"] = detectTool("docker", "--version", `Docker version (\d+\.\d+\.\d+)`)

	// Docker Compose (v2 style: docker compose)
	dockerComposeV2 := detectToolWithArgs("docker", []string{"compose", "version"}, `v?(\d+\.\d+\.\d+)`)
	if dockerComposeV2.Available {
		dockerComposeV2.Name = "docker-compose"
		tools["docker-compose"] = dockerComposeV2
	} else {
		// Fall back to docker-compose (v1)
		tools["docker-compose"] = detectTool("docker-compose", "--version", `docker-compose version (\d+\.\d+\.\d+)`)
	}

	// Go
	tools["go"] = detectTool("go", "version", `go(\d+\.\d+\.?\d*)`)

	// Node
	tools["node"] = detectTool("node", "--version", `v?(\d+\.\d+\.\d+)`)

	// Python
	tools["python"] = detectTool("python3", "--version", `Python (\d+\.\d+\.\d+)`)
	if !tools["python"].Available {
		tools["python"] = detectTool("python", "--version", `Python (\d+\.\d+\.\d+)`)
	}

	// npm
	tools["npm"] = detectTool("npm", "--version", `(\d+\.\d+\.\d+)`)

	// pnpm
	tools["pnpm"] = detectTool("pnpm", "--version", `(\d+\.\d+\.\d+)`)

	// yarn
	tools["yarn"] = detectTool("yarn", "--version", `(\d+\.\d+\.\d+)`)

	// Make
	tools["make"] = detectTool("make", "--version", `GNU Make (\d+\.\d+\.?\d*)`)

	return tools
}

// detectTool detects a tool's version
func detectTool(command, args, pattern string) ToolInfo {
	return detectToolWithArgs(command, strings.Fields(args), pattern)
}

// detectToolWithArgs detects a tool's version with multiple args
func detectToolWithArgs(command string, args []string, pattern string) ToolInfo {
	info := ToolInfo{
		Name: command,
	}

	// Check if command exists
	path, err := exec.LookPath(command)
	if err != nil {
		info.Available = false
		info.Error = "not found in PATH"
		return info
	}

	info.Path = path
	info.Available = true

	// Run command to get version
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		info.Error = fmt.Sprintf("failed to get version: %v", err)
		return info
	}

	// Extract version
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		info.Version = matches[1]
	} else {
		info.Error = "could not parse version"
	}

	return info
}

// CheckVersions checks if tools meet minimum version requirements
func CheckVersions(requirements map[string]string) []VersionCheck {
	tools := DetectTools()
	var results []VersionCheck

	for tool, minVersion := range requirements {
		if minVersion == "" {
			continue
		}

		info, exists := tools[tool]
		check := VersionCheck{
			Tool:     tool,
			Required: minVersion,
		}

		if !exists || !info.Available {
			check.Available = false
			check.Satisfied = false
			if info.Error != "" {
				check.Error = info.Error
			} else {
				check.Error = "tool not found"
			}
			results = append(results, check)
			continue
		}

		check.Available = true
		check.Current = info.Version
		check.Satisfied = CompareVersions(info.Version, minVersion) >= 0

		results = append(results, check)
	}

	return results
}

// CompareVersions compares two semver-like versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	for i := 0; i < 3; i++ {
		p1, p2 := 0, 0
		if i < len(parts1) {
			p1 = parts1[i]
		}
		if i < len(parts2) {
			p2 = parts2[i]
		}

		if p1 < p2 {
			return -1
		}
		if p1 > p2 {
			return 1
		}
	}

	return 0
}

// parseVersion extracts numeric version parts
func parseVersion(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	result := make([]int, 0, len(parts))

	for _, p := range parts {
		// Handle versions like "20.10" that might have extra text
		numStr := strings.TrimFunc(p, func(r rune) bool {
			return r < '0' || r > '9'
		})
		n, _ := strconv.Atoi(numStr)
		result = append(result, n)
	}

	return result
}
