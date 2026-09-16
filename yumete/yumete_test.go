package yumete

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lost-melody/yuman/release"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestParseName(t *testing.T) {
	tests := []struct {
		name        string
		wantVersion string
		wantArch    string
		wantOK      bool
	}{
		{name: "yumete-0.1.0-linux-x86_64", wantVersion: "0.1.0", wantArch: "x86_64", wantOK: true},
		{name: "yumete-0.1.0-linux-aarch64", wantVersion: "0.1.0", wantArch: "aarch64", wantOK: true},
		{name: "yumete-0.1.0-beta.1-linux-x86_64", wantVersion: "0.1.0-beta.1", wantArch: "x86_64", wantOK: true},
		{name: "yumete-0.1.0-linux", wantOK: false},
		{name: "yumete-linux-x86_64", wantOK: false},
		{name: "yumete-0.1.0-darwin-x86_64", wantOK: false},
		{name: "yume-0.1.0-linux-x86_64", wantOK: false},
		{name: "yumete", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, arch, ok := parseName(tt.name)
			if ok != tt.wantOK {
				t.Fatalf("parseName(%q) ok = %v, want %v", tt.name, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if version != tt.wantVersion || arch != tt.wantArch {
				t.Errorf("parseName(%q) = (%q, %q), want (%q, %q)",
					tt.name, version, arch, tt.wantVersion, tt.wantArch)
			}
		})
	}
}

func TestAssetName(t *testing.T) {
	tests := []struct {
		name        string
		wantVersion string
		wantArch    string
		wantOK      bool
	}{
		{name: "yumete-0.1.0-linux-x86_64.tar.gz", wantVersion: "0.1.0", wantArch: "x86_64", wantOK: true},
		{name: "yumete-0.1.0-linux-aarch64.tar.gz", wantVersion: "0.1.0", wantArch: "aarch64", wantOK: true},
		{name: "yumete-0.1.0-linux-x86_64.tar.gz.sha256", wantOK: false},
		{name: "yumete-0.1.0-darwin-arm64.tar.gz", wantOK: false},
		{name: "v0.1.0.tar.gz", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, arch, ok := assetName(tt.name)
			if ok != tt.wantOK {
				t.Fatalf("assetName(%q) ok = %v, want %v", tt.name, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if version != tt.wantVersion || arch != tt.wantArch {
				t.Errorf("assetName(%q) = (%q, %q), want (%q, %q)",
					tt.name, version, arch, tt.wantVersion, tt.wantArch)
			}
		})
	}
}

func TestSelectAsset(t *testing.T) {
	assets := []release.Asset{
		{Name: "yumete-0.1.0-darwin-arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin"},
		{Name: "yumete-0.1.0-linux-aarch64.tar.gz", BrowserDownloadURL: "https://example.com/arm"},
		{Name: "yumete-0.1.0-linux-x86_64.tar.gz", BrowserDownloadURL: "https://example.com/x86"},
		{Name: "yumete-0.1.0-linux-x86_64.tar.gz.sha256", BrowserDownloadURL: "https://example.com/sum"},
		{Name: "v0.1.0.tar.gz", BrowserDownloadURL: "https://example.com/source"},
	}

	got, err := selectAsset(assets, "x86_64")
	if err != nil {
		t.Fatal(err)
	}
	if want := "https://example.com/x86"; got.BrowserDownloadURL != want {
		t.Errorf("selectAsset() URL = %q, want %q", got.BrowserDownloadURL, want)
	}

	if _, err := selectAsset(assets, "riscv64"); err == nil {
		t.Error("selectAsset() = nil error, want no matching asset error")
	}
}

func TestSelectAssetPicksNewest(t *testing.T) {
	assets := []release.Asset{
		{Name: "yumete-0.1.0-linux-x86_64.tar.gz", BrowserDownloadURL: "https://example.com/old"},
		{Name: "yumete-0.2.0-linux-x86_64.tar.gz", BrowserDownloadURL: "https://example.com/new"},
	}

	got, err := selectAsset(assets, "x86_64")
	if err != nil {
		t.Fatal(err)
	}
	if want := "https://example.com/new"; got.BrowserDownloadURL != want {
		t.Errorf("selectAsset() URL = %q, want %q", got.BrowserDownloadURL, want)
	}
}

func TestListPackages(t *testing.T) {
	t.Run("missing lib home", func(t *testing.T) {
		packages, err := listPackages(filepath.Join(t.TempDir(), "absent"))
		if err != nil {
			t.Fatal(err)
		}
		if len(packages) != 0 {
			t.Errorf("listPackages() = %v, want none", packages)
		}
	})

	t.Run("sorted and filtered", func(t *testing.T) {
		libHome := t.TempDir()
		for _, name := range []string{
			"yumete-0.1.0-linux-x86_64",
			"yumete-0.3.0-linux-x86_64",
			"yumete-0.2.0-linux-aarch64",
			"yumete-not-a-package",
			"fcitx5",
		} {
			if err := os.Mkdir(filepath.Join(libHome, name), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		// A file and a symlink named like a package must be ignored.
		writeFile(t, filepath.Join(libHome, "yumete-9.9.9-linux-x86_64"), "")
		if err := os.Symlink(filepath.Join(libHome, "fcitx5"), filepath.Join(libHome, "yumete-9.8.8-linux-x86_64")); err != nil {
			t.Fatal(err)
		}

		packages, err := listPackages(libHome)
		if err != nil {
			t.Fatal(err)
		}
		got := make([]string, len(packages))
		for i, pkg := range packages {
			got[i] = pkg.Name
			if pkg.Path != filepath.Join(libHome, pkg.Name) {
				t.Errorf("package %q path = %q, want it under %q", pkg.Name, pkg.Path, libHome)
			}
		}
		want := []string{
			"yumete-0.3.0-linux-x86_64",
			"yumete-0.2.0-linux-aarch64",
			"yumete-0.1.0-linux-x86_64",
		}
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("listPackages() = %v, want %v", got, want)
		}
		if packages[1].Version != "0.2.0" || packages[1].Arch != "aarch64" {
			t.Errorf("package[1] = %+v, want version 0.2.0 on aarch64", packages[1])
		}
	})
}

// writeTarball builds a yumete-<version>-linux-x86_64.tar.gz holding the same
// entries as a real release, with the binary at the given mode.
func writeTarball(t *testing.T, dir, packageDir string, binaryMode int64) string {
	t.Helper()

	path := filepath.Join(dir, packageDir+".tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	entries := []struct {
		name    string
		mode    int64
		content string
	}{
		{packageDir + "/", 0o755, ""},
		{packageDir + "/LICENSE", 0o644, "license"},
		{packageDir + "/bin/", 0o755, ""},
		{packageDir + "/bin/yumete", binaryMode, "#!/bin/sh\necho yumete\n"},
		{packageDir + "/docs/", 0o755, ""},
		{packageDir + "/docs/manual.md", 0o644, "manual"},
	}
	for _, entry := range entries {
		hdr := &tar.Header{Name: entry.name, Mode: entry.mode, Size: int64(len(entry.content))}
		if strings.HasSuffix(entry.name, "/") {
			hdr.Typeflag = tar.TypeDir
			hdr.Size = 0
		} else {
			hdr.Typeflag = tar.TypeReg
		}
		if err = tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if hdr.Size > 0 {
			if _, err = tw.Write([]byte(entry.content)); err != nil {
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

func TestStagedPackage(t *testing.T) {
	newStaging := func(t *testing.T, dirs ...string) string {
		t.Helper()
		staging := t.TempDir()
		for _, dir := range dirs {
			if err := os.MkdirAll(filepath.Join(staging, dir), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		return staging
	}

	t.Run("valid package", func(t *testing.T) {
		const name = "yumete-0.1.0-linux-x86_64"
		staging := newStaging(t, filepath.Join(name, "bin"))
		writeFile(t, filepath.Join(staging, name, "bin", "yumete"), "")

		gotName, dir, err := stagedPackage(staging)
		if err != nil {
			t.Fatal(err)
		}
		if gotName != name || dir != filepath.Join(staging, name) {
			t.Errorf("stagedPackage() = (%q, %q), want (%q, %q)", gotName, dir, name, filepath.Join(staging, name))
		}
	})

	t.Run("no package directory", func(t *testing.T) {
		if _, _, err := stagedPackage(newStaging(t)); err == nil {
			t.Error("stagedPackage() = nil error, want layout error")
		}
	})

	t.Run("several package directories", func(t *testing.T) {
		staging := newStaging(t, "yumete-0.1.0-linux-x86_64", "yumete-0.2.0-linux-x86_64")
		if _, _, err := stagedPackage(staging); err == nil {
			t.Error("stagedPackage() = nil error, want layout error")
		}
	})

	t.Run("unexpected directory name", func(t *testing.T) {
		if _, _, err := stagedPackage(newStaging(t, "release")); err == nil {
			t.Error("stagedPackage() = nil error, want layout error")
		}
	})

	t.Run("missing binary", func(t *testing.T) {
		staging := newStaging(t, "yumete-0.1.0-linux-x86_64")
		if _, _, err := stagedPackage(staging); err == nil {
			t.Error("stagedPackage() = nil error, want missing binary error")
		}
	})
}

func TestReplaceDir(t *testing.T) {
	t.Run("moves into place", func(t *testing.T) {
		root := t.TempDir()
		src := filepath.Join(root, "staging", "pkg")
		writeFile(t, filepath.Join(src, "bin", "yumete"), "")
		target := filepath.Join(root, "lib", "yumete-0.1.0-linux-x86_64")

		if err := replaceDir(src, target); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(target, "bin", "yumete")); err != nil {
			t.Errorf("target should hold the moved package: %v", err)
		}
	})

	t.Run("replaces an existing target", func(t *testing.T) {
		root := t.TempDir()
		src := filepath.Join(root, "staging", "pkg")
		writeFile(t, filepath.Join(src, "bin", "yumete"), "")
		target := filepath.Join(root, "lib", "yumete-0.1.0-linux-x86_64")
		writeFile(t, filepath.Join(target, "stale"), "")

		if err := replaceDir(src, target); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(target, "bin", "yumete")); err != nil {
			t.Errorf("target should hold the moved package: %v", err)
		}
		if _, err := os.Stat(filepath.Join(target, "stale")); !os.IsNotExist(err) {
			t.Errorf("stale entry should be gone, stat err = %v", err)
		}
	})
}

func TestLinkBinary(t *testing.T) {
	const name = "yumete-0.1.0-linux-x86_64"

	t.Run("creates a relative link", func(t *testing.T) {
		root := t.TempDir()
		binHome := filepath.Join(root, ".local", "bin")
		if err := linkBinary(binHome, name); err != nil {
			t.Fatal(err)
		}

		link := filepath.Join(binHome, "yumete")
		target, err := os.Readlink(link)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join("..", "lib", name, "bin", "yumete")
		if target != want {
			t.Errorf("symlink target = %q, want %q", target, want)
		}
	})

	t.Run("replaces an existing link", func(t *testing.T) {
		root := t.TempDir()
		binHome := filepath.Join(root, ".local", "bin")
		if err := os.MkdirAll(binHome, 0o755); err != nil {
			t.Fatal(err)
		}
		old := "yumete-0.0.9-linux-x86_64"
		if err := os.Symlink(filepath.Join("..", "lib", old, "bin", "yumete"), filepath.Join(binHome, "yumete")); err != nil {
			t.Fatal(err)
		}

		if err := linkBinary(binHome, name); err != nil {
			t.Fatal(err)
		}
		target, err := os.Readlink(filepath.Join(binHome, "yumete"))
		if err != nil {
			t.Fatal(err)
		}
		if want := filepath.Join("..", "lib", name, "bin", "yumete"); target != want {
			t.Errorf("symlink target = %q, want %q", target, want)
		}
		if _, err := os.Lstat(filepath.Join(binHome, "yumete.tmp")); !os.IsNotExist(err) {
			t.Errorf("temporary link should be gone, lstat err = %v", err)
		}
	})
}

func TestRemoveReplacedPackages(t *testing.T) {
	newPackage := func(t *testing.T, libHome, name string) Package {
		t.Helper()
		path := filepath.Join(libHome, name)
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
		version, arch, ok := parseName(name)
		if !ok {
			t.Fatalf("test package name %q does not parse", name)
		}
		return Package{Name: name, Path: path, Version: version, Arch: arch}
	}

	t.Run("nothing replaced", func(t *testing.T) {
		if err := removeReplacedPackages(nil, true); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("single package removed", func(t *testing.T) {
		libHome := t.TempDir()
		old := newPackage(t, libHome, "yumete-0.0.9-linux-x86_64")

		if err := removeReplacedPackages([]Package{old}, true); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(old.Path); !os.IsNotExist(err) {
			t.Errorf("old package should be removed, stat err = %v", err)
		}
	})

	t.Run("several packages kept", func(t *testing.T) {
		libHome := t.TempDir()
		first := newPackage(t, libHome, "yumete-0.0.8-linux-x86_64")
		second := newPackage(t, libHome, "yumete-0.0.9-linux-x86_64")

		if err := removeReplacedPackages([]Package{second, first}, false); err != nil {
			t.Fatal(err)
		}
		for _, pkg := range []Package{first, second} {
			if _, err := os.Stat(pkg.Path); err != nil {
				t.Errorf("package %q should be kept: %v", pkg.Name, err)
			}
		}
	})
}

func TestInstallPackage(t *testing.T) {
	const name = "yumete-0.1.0-linux-x86_64"
	latest := Release{Version: "0.1.0", Arch: "x86_64"}

	newLayout := func(t *testing.T) layout {
		t.Helper()
		root := t.TempDir()
		return layout{
			libHome: filepath.Join(root, ".local", "lib"),
			binHome: filepath.Join(root, ".local", "bin"),
		}
	}

	assertInstalled := func(t *testing.T, dirs layout) {
		t.Helper()
		binary := filepath.Join(dirs.libHome, name, "bin", "yumete")
		info, err := os.Stat(binary)
		if err != nil {
			t.Fatalf("installed binary is missing: %v", err)
		}
		if got := info.Mode().Perm(); got != 0o755 {
			t.Errorf("binary mode = %v, want 0755", got)
		}
		if _, err = os.Stat(filepath.Join(dirs.libHome, name, "docs", "manual.md")); err != nil {
			t.Errorf("package docs were not extracted: %v", err)
		}

		link := filepath.Join(dirs.binHome, "yumete")
		resolved, err := filepath.EvalSymlinks(link)
		if err != nil {
			t.Fatalf("symlink does not resolve: %v", err)
		}
		if resolved != binary {
			t.Errorf("symlink resolves to %q, want %q", resolved, binary)
		}

		// The staging directory must not survive the install.
		entries, err := os.ReadDir(dirs.libHome)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".yumete-install-") {
				t.Errorf("staging directory %q was left behind", entry.Name())
			}
		}
	}

	t.Run("fresh install", func(t *testing.T) {
		dirs := newLayout(t)
		// A release binary that lost its executable bit must be fixed up.
		pkg := writeTarball(t, t.TempDir(), name, 0o644)

		if err := installPackage(context.Background(), dirs, latest, pkg, nil, false); err != nil {
			t.Fatal(err)
		}
		assertInstalled(t, dirs)
	})

	t.Run("update removes the replaced package", func(t *testing.T) {
		dirs := newLayout(t)
		old := Package{
			Name:    "yumete-0.0.9-linux-x86_64",
			Path:    filepath.Join(dirs.libHome, "yumete-0.0.9-linux-x86_64"),
			Version: "0.0.9",
			Arch:    "x86_64",
		}
		if err := os.MkdirAll(old.Path, 0o755); err != nil {
			t.Fatal(err)
		}
		pkg := writeTarball(t, t.TempDir(), name, 0o755)

		if err := installPackage(context.Background(), dirs, latest, pkg, []Package{old}, true); err != nil {
			t.Fatal(err)
		}
		assertInstalled(t, dirs)
		if _, err := os.Stat(old.Path); !os.IsNotExist(err) {
			t.Errorf("replaced package should be removed, stat err = %v", err)
		}
	})

	t.Run("update keeps several old packages", func(t *testing.T) {
		dirs := newLayout(t)
		var old []Package
		for _, version := range []string{"0.0.8", "0.0.9"} {
			pkg := Package{
				Name:    "yumete-" + version + "-linux-x86_64",
				Path:    filepath.Join(dirs.libHome, "yumete-"+version+"-linux-x86_64"),
				Version: version,
				Arch:    "x86_64",
			}
			if err := os.MkdirAll(pkg.Path, 0o755); err != nil {
				t.Fatal(err)
			}
			old = append(old, pkg)
		}
		pkg := writeTarball(t, t.TempDir(), name, 0o755)

		if err := installPackage(context.Background(), dirs, latest, pkg, old, false); err != nil {
			t.Fatal(err)
		}
		assertInstalled(t, dirs)
		for _, replaced := range old {
			if _, err := os.Stat(replaced.Path); err != nil {
				t.Errorf("package %q should be kept: %v", replaced.Name, err)
			}
		}
	})

	t.Run("rejects a package without the expected layout", func(t *testing.T) {
		dirs := newLayout(t)
		pkg := writeTarball(t, t.TempDir(), "not-yumete-0.1.0-linux-x86_64", 0o755)

		if err := installPackage(context.Background(), dirs, latest, pkg, nil, false); err == nil {
			t.Error("installPackage() = nil error, want layout error")
		}
	})
}
