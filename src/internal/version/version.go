package version

// Build information set via ldflags at compile time
var (
	// Commit is the git commit hash
	Commit = "unknown"
	// BuildTime is the time the binary was built
	BuildTime = "unknown"
	// Version is the semantic version (if tagged)
	Version = "dev"
)

// Info contains build information
type Info struct {
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
	Version   string `json:"version"`
}

// Get returns the current build information
func Get() Info {
	return Info{
		Commit:    Commit,
		BuildTime: BuildTime,
		Version:   Version,
	}
}
