package models

// ArtifactType represents the type of artifact detected
type ArtifactType string

const (
	ArtifactCompose    ArtifactType = "compose"
	ArtifactEnv        ArtifactType = "env"
	ArtifactEnvExample ArtifactType = "env_example"
	ArtifactManifest   ArtifactType = "manifest"
	ArtifactReadme     ArtifactType = "readme"
	ArtifactMakefile   ArtifactType = "makefile"
)

// Language represents detected programming language
type Language string

const (
	LangNodeJS Language = "nodejs"
	LangGo     Language = "go"
	LangPython Language = "python"
	LangRust   Language = "rust"
	LangJava   Language = "java"
	LangCSharp Language = "csharp"
	LangUnknown Language = "unknown"
)

// Artifact represents a detected file or configuration
type Artifact struct {
	Type     ArtifactType `json:"type"`
	Path     string       `json:"path"`
	Language Language     `json:"language,omitempty"`
	Details  string       `json:"details,omitempty"`
	Found    bool         `json:"found"`
}

// Artifacts is a collection of detected artifacts
type Artifacts struct {
	ComposeFiles   []Artifact `json:"compose_files"`
	EnvFiles       []Artifact `json:"env_files"`
	EnvExamples    []Artifact `json:"env_examples"`
	Manifests      []Artifact `json:"manifests"`
	Readme         *Artifact  `json:"readme,omitempty"`
	Makefile       *Artifact  `json:"makefile,omitempty"`
	DetectedLang   Language   `json:"detected_language,omitempty"`
	PackageManager string     `json:"package_manager,omitempty"`
}

// NewArtifacts creates a new empty Artifacts
func NewArtifacts() *Artifacts {
	return &Artifacts{
		ComposeFiles: make([]Artifact, 0),
		EnvFiles:     make([]Artifact, 0),
		EnvExamples:  make([]Artifact, 0),
		Manifests:    make([]Artifact, 0),
	}
}

// HasCompose checks if any compose file was found
func (a *Artifacts) HasCompose() bool {
	for _, c := range a.ComposeFiles {
		if c.Found {
			return true
		}
	}
	return false
}

// HasEnv checks if any .env file was found
func (a *Artifacts) HasEnv() bool {
	for _, e := range a.EnvFiles {
		if e.Found {
			return true
		}
	}
	return false
}

// HasEnvExample checks if any .env.example file was found
func (a *Artifacts) HasEnvExample() bool {
	for _, e := range a.EnvExamples {
		if e.Found {
			return true
		}
	}
	return false
}
