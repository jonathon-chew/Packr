package registry

type Package struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Homepage    string    `yaml:"homepage"`
	GitHub      *GitHub   `yaml:"github,omitempty"`
	Releases    []Release `yaml:"releases"`
}

type GitHub struct {
	Owner  string `yaml:"owner"`
	Repo   string `yaml:"repo"`
	Binary string `yaml:"binary,omitempty"`
}

type Release struct {
	Version string            `yaml:"version"`
	Targets map[string]Target `yaml:"targets"` // key: "linux-amd64", etc.
}

type Target struct {
	URL         string `yaml:"url"`
	SHA256      string `yaml:"sha256"`
	Bin         string `yaml:"bin"`          // path to the binary inside the archive
	ArchiveType string `yaml:"archive_type"` // optional override: tar.gz, zip, binary
}
