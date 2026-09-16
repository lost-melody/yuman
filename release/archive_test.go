package release

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// entry is one file or directory to put into a test tarball.
type entry struct {
	name    string
	mode    int64
	content string
	dir     bool
}

// writeTarball builds a .tar.gz holding entries and returns its path.
func writeTarball(t *testing.T, entries []entry) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "package.tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	for _, e := range entries {
		hdr := &tar.Header{Name: e.name, Mode: e.mode, Size: int64(len(e.content))}
		if e.dir {
			hdr.Typeflag = tar.TypeDir
			hdr.Size = 0
		} else {
			hdr.Typeflag = tar.TypeReg
		}
		if err = tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if hdr.Size > 0 {
			if _, err = tw.Write([]byte(e.content)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err = tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err = gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractTarGz(t *testing.T) {
	pkg := writeTarball(t, []entry{
		{name: "pkg/", mode: 0o755, dir: true},
		{name: "pkg/bin/", mode: 0o755, dir: true},
		{name: "pkg/bin/tool", mode: 0o755, content: "#!/bin/sh\n"},
		{name: "pkg/README.md", mode: 0o644, content: "readme"},
	})

	destDir := t.TempDir()
	if err := ExtractTarGz(context.Background(), pkg, destDir); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(destDir, "pkg", "bin", "tool"))
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o755 {
		t.Errorf("extracted mode = %v, want 0755", got)
	}
	data, err := os.ReadFile(filepath.Join(destDir, "pkg", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "readme" {
		t.Errorf("extracted content = %q, want %q", data, "readme")
	}
}

func TestExtractTarGzRejectsBadEntries(t *testing.T) {
	tests := []struct {
		name  string
		entry entry
	}{
		{name: "absolute path", entry: entry{name: "/etc/passwd", mode: 0o644, content: "x"}},
		{name: "escaping path", entry: entry{name: "../escape", mode: 0o644, content: "x"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := writeTarball(t, []entry{tt.entry})
			if err := ExtractTarGz(context.Background(), pkg, t.TempDir()); err == nil {
				t.Error("ExtractTarGz() = nil error, want rejection")
			}
		})
	}
}

func TestExtractTarGzStopsOnCanceledContext(t *testing.T) {
	pkg := writeTarball(t, []entry{
		{name: "pkg/", mode: 0o755, dir: true},
		{name: "pkg/file", mode: 0o644, content: "x"},
	})

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	if err := ExtractTarGz(ctx, pkg, t.TempDir()); err == nil {
		t.Error("ExtractTarGz() = nil error, want context error")
	}
}
