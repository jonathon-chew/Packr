package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindBinaryUsesNamedMatch(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "tool", "rg")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := findBinary(root, "rg")
	if err != nil {
		t.Fatal(err)
	}
	if got != path {
		t.Fatalf("expected %q, got %q", path, got)
	}
}

func TestFindBinaryFallsBackToExecutableCandidate(t *testing.T) {
	root := t.TempDir()
	readme := filepath.Join(root, "README.md")
	bin := filepath.Join(root, "bin", "tool")

	if err := os.WriteFile(readme, []byte("docs"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bin, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := findBinary(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if got != bin {
		t.Fatalf("expected %q, got %q", bin, got)
	}
}
