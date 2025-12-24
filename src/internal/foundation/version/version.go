package version

// Build information set via ldflags at compile time
var (
	// Commit is the git commit hash
	Commit = "unknown"
	// BuildTime is the time the binary was built
	BuildTime = "unknown"
)

// Info contains build information
type Info struct {
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

// Get returns the current build information
func Get() Info {
	return Info{
		Commit:    Commit,
		BuildTime: BuildTime,
	}
}
