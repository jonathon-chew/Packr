package registry

type Package struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Homepage    string    `yaml:"homepage"`
	Releases    []Release `yaml:"releases"`
}

type Release struct {
	Version string            `yaml:"version"`
	Targets map[string]Target `yaml:"targets"` // key: "linux-amd64", etc.
}

type Target struct {
	URL    string `yaml:"url"`
	SHA256 string `yaml:"sha256"`
	Bin    string `yaml:"bin"` // path to the binary inside the archive
}
