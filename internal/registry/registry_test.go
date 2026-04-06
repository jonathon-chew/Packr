package registry

import "testing"

func TestChooseBestAssetPrefersMatchingArchive(t *testing.T) {
	assets := []githubAssetBrief{
		{Name: "tool_1.0.0_checksums.txt"},
		{Name: "tool_1.0.0_x86_64-unknown-linux-gnu.tar.gz", BrowserDownloadURL: "gnu"},
		{Name: "tool_1.0.0_x86_64-unknown-linux-musl.tar.gz", BrowserDownloadURL: "musl"},
		{Name: "tool_1.0.0_aarch64-apple-darwin.tar.gz", BrowserDownloadURL: "darwin"},
	}

	asset, ok := chooseBestAsset(assets, "linux", "amd64")
	if !ok {
		t.Fatal("expected a matching asset")
	}
	if asset.BrowserDownloadURL != "musl" {
		t.Fatalf("expected musl asset, got %q", asset.BrowserDownloadURL)
	}
}

func TestChooseBestAssetSkipsSignatures(t *testing.T) {
	assets := []githubAssetBrief{
		{Name: "tool_darwin_arm64.tar.gz.sig"},
		{Name: "tool_darwin_arm64.zip", BrowserDownloadURL: "zip"},
	}

	asset, ok := chooseBestAsset(assets, "darwin", "arm64")
	if !ok {
		t.Fatal("expected a matching asset")
	}
	if asset.BrowserDownloadURL != "zip" {
		t.Fatalf("expected zip asset, got %q", asset.BrowserDownloadURL)
	}
}
