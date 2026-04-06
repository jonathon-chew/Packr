package registry

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed *.yaml
var bundledRegistry embed.FS

// LoadPackage resolves a package from a user registry directory, bundled YAML,
// or directly from a GitHub repository reference like "owner/repo".
func LoadPackage(registryDir, name, version string) (Package, Release, error) {
	pkg, err := loadPackageDefinition(registryDir, name)
	switch {
	case err == nil:
		release, resolveErr := ResolveRelease(pkg, version)
		if resolveErr != nil {
			return Package{}, Release{}, resolveErr
		}
		return pkg, release, nil
	case !os.IsNotExist(err):
		return Package{}, Release{}, err
	case strings.Count(name, "/") == 1:
		return LoadGitHubPackage(name, version)
	default:
		return Package{}, Release{}, err
	}
}

func loadPackageDefinition(registryDir, name string) (Package, error) {
	data, err := loadPackageYAML(registryDir, name)
	if err != nil {
		return Package{}, err
	}

	var pkg Package
	if err := yaml.Unmarshal(data, &pkg); err != nil {
		return Package{}, fmt.Errorf("parsing package definition for %q: %w", name, err)
	}

	return pkg, nil
}

func loadPackageYAML(registryDir, name string) ([]byte, error) {
	filename := name + ".yaml"
	if registryDir != "" {
		path := filepath.Join(registryDir, filename)
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
	}

	data, err := bundledRegistry.ReadFile(filename)
	if err == nil {
		return data, nil
	}
	if errorsIsNotExist(err) {
		return nil, os.ErrNotExist
	}
	return nil, fmt.Errorf("reading bundled registry entry %q: %w", filename, err)
}

// ResolveRelease picks a release by version; if version == "", returns the first.
func ResolveRelease(pkg Package, version string) (Release, error) {
	if pkg.GitHub != nil {
		ref := pkg.GitHub.Owner + "/" + pkg.GitHub.Repo
		return resolveGitHubRelease(*pkg.GitHub, ref, version)
	}

	if len(pkg.Releases) == 0 {
		return Release{}, fmt.Errorf("no releases defined")
	}

	if version == "" {
		return pkg.Releases[0], nil // later you can sort or do semver
	}

	for _, r := range pkg.Releases {
		if r.Version == version {
			return r, nil
		}
	}

	return Release{}, fmt.Errorf("version %q not found", version)
}

func LoadGitHubPackage(repoRef, version string) (Package, Release, error) {
	parts := strings.SplitN(repoRef, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return Package{}, Release{}, fmt.Errorf("invalid GitHub repository %q; use owner/repo", repoRef)
	}

	pkg := Package{
		Name:     parts[1],
		Homepage: "https://github.com/" + repoRef,
		GitHub: &GitHub{
			Owner: parts[0],
			Repo:  parts[1],
		},
	}

	release, err := resolveGitHubRelease(*pkg.GitHub, repoRef, version)
	if err != nil {
		return Package{}, Release{}, err
	}
	return pkg, release, nil
}

func resolveGitHubRelease(gh GitHub, repoRef, version string) (Release, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repoRef)
	if version != "" {
		endpoint = fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repoRef, version)
	}

	releaseInfo, err := fetchGitHubRelease(endpoint)
	if err != nil {
		return Release{}, err
	}

	target, err := selectGitHubTarget(gh, repoRef, releaseInfo)
	if err != nil {
		return Release{}, err
	}

	return Release{
		Version: releaseInfo.TagName,
		Targets: map[string]Target{
			platformKey(): target,
		},
	}, nil
}

type githubReleaseResponse struct {
	TagName string             `json:"tag_name"`
	Assets  []githubAssetBrief `json:"assets"`
}

type githubAssetBrief struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func fetchGitHubRelease(endpoint string) (githubReleaseResponse, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return githubReleaseResponse{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "packr")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return githubReleaseResponse{}, fmt.Errorf("requesting GitHub release metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return githubReleaseResponse{}, fmt.Errorf("GitHub API %s returned %d: %s", endpoint, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var release githubReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return githubReleaseResponse{}, fmt.Errorf("decoding GitHub release metadata: %w", err)
	}
	if release.TagName == "" {
		return githubReleaseResponse{}, fmt.Errorf("GitHub release metadata did not include a tag name")
	}
	return release, nil
}

func selectGitHubTarget(gh GitHub, repoRef string, release githubReleaseResponse) (Target, error) {
	pk := platformKey()
	asset, ok := chooseBestAsset(release.Assets, runtime.GOOS, runtime.GOARCH)
	if !ok {
		return Target{}, fmt.Errorf("no release asset matched %s for %s@%s", pk, repoRef, release.TagName)
	}

	bin := gh.Binary
	if bin == "" {
		bin = strings.TrimSuffix(gh.Repo, filepath.Ext(gh.Repo))
	}

	return Target{
		URL:         asset.BrowserDownloadURL,
		Bin:         bin,
		ArchiveType: archiveTypeForAsset(asset.Name),
	}, nil
}

func chooseBestAsset(assets []githubAssetBrief, goos, goarch string) (githubAssetBrief, bool) {
	type scored struct {
		asset githubAssetBrief
		score int
	}

	var best scored
	found := false
	for _, asset := range assets {
		score, ok := scoreAsset(asset.Name, goos, goarch)
		if !ok {
			continue
		}
		if !found || score > best.score {
			best = scored{asset: asset, score: score}
			found = true
		}
	}

	return best.asset, found
}

func scoreAsset(name, goos, goarch string) (int, bool) {
	lower := strings.ToLower(name)

	for _, blocked := range []string{
		".sha256", ".sha512", ".sig", ".minisig", ".pem", ".sbom", "checksums",
		"sha256sum", "sha512sum", "provenance", "attestation",
	} {
		if strings.Contains(lower, blocked) {
			return 0, false
		}
	}

	score := 0
	if hasAnyToken(lower, osTokens(goos)) {
		score += 4
	} else {
		return 0, false
	}

	if hasAnyToken(lower, archTokens(goarch)) {
		score += 4
	} else {
		return 0, false
	}

	switch archiveTypeForAsset(lower) {
	case "tar.gz":
		score += 3
	case "zip":
		score += 2
	case "binary":
		score += 1
	default:
		return 0, false
	}

	if strings.Contains(lower, "musl") {
		score++
	}
	if strings.Contains(lower, "static") {
		score++
	}
	if strings.Contains(lower, "gnu") || strings.Contains(lower, "msvc") {
		score--
	}

	return score, true
}

func archiveTypeForAsset(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return "tar.gz"
	case strings.HasSuffix(lower, ".zip"):
		return "zip"
	case !strings.Contains(filepath.Base(lower), "."):
		return "binary"
	default:
		return "unknown"
	}
}

func osTokens(goos string) []string {
	switch goos {
	case "darwin":
		return []string{"darwin", "apple-darwin", "macos", "mac", "osx"}
	case "linux":
		return []string{"linux", "unknown-linux", "linux-musl", "linux-gnu"}
	case "windows":
		return []string{"windows", "win"}
	default:
		return []string{goos}
	}
}

func archTokens(goarch string) []string {
	switch goarch {
	case "amd64":
		return []string{"x86_64", "amd64", "x64", "64bit"}
	case "arm64":
		return []string{"aarch64", "arm64"}
	default:
		return []string{goarch}
	}
}

func hasAnyToken(name string, tokens []string) bool {
	for _, token := range tokens {
		if strings.Contains(name, token) {
			return true
		}
	}
	return false
}

func platformKey() string {
	return strings.Join([]string{runtime.GOOS, runtime.GOARCH}, "-")
}

func errorsIsNotExist(err error) bool {
	return err != nil && (os.IsNotExist(err) || errors.Is(err, fs.ErrNotExist))
}
