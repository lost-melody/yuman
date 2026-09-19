package selfupdate

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
		name          string
		wantVersion   string
		wantTimestamp string
		wantArch      string
		wantOK        bool
	}{
		{name: "yuman-v0.2.0-202609181503-x86_64", wantVersion: "v0.2.0", wantTimestamp: "202609181503", wantArch: "x86_64", wantOK: true},
		{name: "yuman-v0.2.0-202609181503-aarch64", wantVersion: "v0.2.0", wantTimestamp: "202609181503", wantArch: "aarch64", wantOK: true},
		{name: "yuman-v0.2.0-beta.1-202609181503-x86_64", wantVersion: "v0.2.0-beta.1", wantTimestamp: "202609181503", wantArch: "x86_64", wantOK: true},
		{name: "yuman-0.2.0-202609181503-x86_64", wantVersion: "0.2.0", wantTimestamp: "202609181503", wantArch: "x86_64", wantOK: true},
		{name: "yuman-v0.2.0-x86_64", wantOK: false},
		{name: "yuman-202609181503-x86_64", wantOK: false},
		{name: "yuman-v0.2.0-202609181503", wantOK: false},
		{name: "yumete-0.1.0-linux-x86_64", wantOK: false},
		{name: "yuman", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, timestamp, arch, ok := parseName(tt.name)
			if ok != tt.wantOK {
				t.Fatalf("parseName(%q) ok = %v, want %v", tt.name, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if version != tt.wantVersion || timestamp != tt.wantTimestamp || arch != tt.wantArch {
				t.Errorf("parseName(%q) = (%q, %q, %q), want (%q, %q, %q)",
					tt.name, version, timestamp, arch, tt.wantVersion, tt.wantTimestamp, tt.wantArch)
			}
		})
	}
}

func TestAssetName(t *testing.T) {
	tests := []struct {
		name          string
		wantVersion   string
		wantTimestamp string
		wantOK        bool
	}{
		{name: "yuman-v0.2.0-202609181503-x86_64.tar.gz", wantVersion: "v0.2.0", wantTimestamp: "202609181503", wantOK: true},
		{name: "yuman-v0.2.0-202609181503-aarch64.tar.gz", wantVersion: "v0.2.0", wantTimestamp: "202609181503", wantOK: true},
		{name: "yuman-v0.2.0-202609181503-x86_64.tar.gz.sha256", wantOK: false},
		{name: "v0.2.0.tar.gz", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, timestamp, _, ok := assetName(tt.name)
			if ok != tt.wantOK {
				t.Fatalf("assetName(%q) ok = %v, want %v", tt.name, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if version != tt.wantVersion || timestamp != tt.wantTimestamp {
				t.Errorf("assetName(%q) = (%q, %q), want (%q, %q)",
					tt.name, version, timestamp, tt.wantVersion, tt.wantTimestamp)
			}
		})
	}
}

func TestSelectAsset(t *testing.T) {
	assets := []release.Asset{
		{Name: "yuman-v0.2.0-202609181503-aarch64.tar.gz", BrowserDownloadURL: "https://example.com/arm"},
		{Name: "yuman-v0.2.0-202609181503-x86_64.tar.gz", BrowserDownloadURL: "https://example.com/x86"},
		{Name: "yuman-v0.2.0-202609181503-x86_64.tar.gz.sha256", BrowserDownloadURL: "https://example.com/sum"},
		{Name: "v0.2.0.tar.gz", BrowserDownloadURL: "https://example.com/source"},
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

func TestSelectAssetPrefersVersionThenTimestamp(t *testing.T) {
	assets := []release.Asset{
		{Name: "yuman-v0.1.0-202609181600-x86_64.tar.gz", BrowserDownloadURL: "https://example.com/older"},
		{Name: "yuman-v0.2.0-202609181500-x86_64.tar.gz", BrowserDownloadURL: "https://example.com/newer"},
		{Name: "yuman-v0.2.0-202609181503-x86_64.tar.gz", BrowserDownloadURL: "https://example.com/newest"},
	}

	got, err := selectAsset(assets, "x86_64")
	if err != nil {
		t.Fatal(err)
	}
	if want := "https://example.com/newest"; got.BrowserDownloadURL != want {
		t.Errorf("selectAsset() URL = %q, want %q", got.BrowserDownloadURL, want)
	}
}

func TestCompareReleases(t *testing.T) {
	latest := Release{Version: "v0.2.0", Timestamp: "202609181503"}
	tests := []struct {
		name      string
		installed Package
		want      int
	}{
		{name: "older version", installed: Package{Version: "v0.1.0", Timestamp: "202609181600"}, want: -1},
		{name: "newer version", installed: Package{Version: "v0.3.0", Timestamp: "202609181500"}, want: 1},
		{name: "same version older build", installed: Package{Version: "v0.2.0", Timestamp: "202609181500"}, want: -1},
		{name: "same version newer build", installed: Package{Version: "v0.2.0", Timestamp: "202609181600"}, want: 1},
		{name: "identical", installed: Package{Version: "v0.2.0", Timestamp: "202609181503"}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareReleases(tt.installed, latest); got != tt.want {
				t.Errorf("CompareReleases() = %d, want %d", got, tt.want)
			}
		})
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
			"yuman-v0.1.0-202609181503-x86_64",
			"yuman-v0.3.0-202609181503-x86_64",
			"yuman-v0.3.0-202609181600-x86_64",
			"yuman-not-a-package",
			"fcitx5",
		} {
			if err := os.Mkdir(filepath.Join(libHome, name), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		// A file and a symlink named like a package must be ignored.
		writeFile(t, filepath.Join(libHome, "yuman-v9.9.9-202609181503-x86_64"), "")
		if err := os.Symlink(filepath.Join(libHome, "fcitx5"), filepath.Join(libHome, "yuman-v9.8.8-202609181503-x86_64")); err != nil {
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
			"yuman-v0.3.0-202609181600-x86_64",
			"yuman-v0.3.0-202609181503-x86_64",
			"yuman-v0.1.0-202609181503-x86_64",
		}
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("listPackages() = %v, want %v", got, want)
		}
		if packages[0].Timestamp != "202609181600" || packages[0].Version != "v0.3.0" {
			t.Errorf("package[0] = %+v, want version v0.3.0 on the newer build", packages[0])
		}
	})
}

func TestPackageForExecutable(t *testing.T) {
	const name = "yuman-v0.1.0-202609181503-x86_64"

	t.Run("inside a package", func(t *testing.T) {
		root := t.TempDir()
		libHome := filepath.Join(root, ".local", "lib")
		binary := filepath.Join(libHome, name, binaryName)
		writeFile(t, binary, "")

		got, ok := packageForExecutable(binary, libHome)
		if !ok {
			t.Fatal("packageForExecutable() ok = false, want true")
		}
		if got.Name != name || got.Path != filepath.Join(libHome, name) || got.Version != "v0.1.0" {
			t.Errorf("packageForExecutable() = %+v, want package %q", got, name)
		}
	})

	t.Run("reached through the managed link", func(t *testing.T) {
		root := t.TempDir()
		libHome := filepath.Join(root, ".local", "lib")
		binHome := filepath.Join(root, ".local", "bin")
		writeFile(t, filepath.Join(libHome, name, binaryName), "")
		if err := linkBinary(binHome, name); err != nil {
			t.Fatal(err)
		}

		got, ok := packageForExecutable(filepath.Join(binHome, binaryName), libHome)
		if !ok {
			t.Fatal("packageForExecutable() ok = false, want true")
		}
		if got.Path != filepath.Join(libHome, name) {
			t.Errorf("packageForExecutable() path = %q, want %q", got.Path, filepath.Join(libHome, name))
		}
	})

	t.Run("outside the lib home is a first install", func(t *testing.T) {
		root := t.TempDir()
		binary := filepath.Join(root, "build", binaryName)
		writeFile(t, binary, "")

		if _, ok := packageForExecutable(binary, filepath.Join(root, ".local", "lib")); ok {
			t.Error("packageForExecutable() ok = true, want false")
		}
	})

	t.Run("missing directory is a first install", func(t *testing.T) {
		libHome := t.TempDir()
		binary := filepath.Join(libHome, name, binaryName)

		if _, ok := packageForExecutable(binary, libHome); ok {
			t.Error("packageForExecutable() ok = true, want false")
		}
	})
}

func TestIsManaged(t *testing.T) {
	const name = "yuman-v0.1.0-202609181503-x86_64"

	t.Run("link points at the running executable", func(t *testing.T) {
		root := t.TempDir()
		binHome := filepath.Join(root, ".local", "bin")
		libHome := filepath.Join(root, ".local", "lib")
		binary := filepath.Join(libHome, name, binaryName)
		writeFile(t, binary, "")
		if err := linkBinary(binHome, name); err != nil {
			t.Fatal(err)
		}

		if !isManaged(binHome, binary) {
			t.Error("isManaged() = false, want true")
		}
	})

	t.Run("link missing or elsewhere", func(t *testing.T) {
		root := t.TempDir()
		binHome := filepath.Join(root, ".local", "bin")
		writeFile(t, filepath.Join(root, "build", binaryName), "")

		if isManaged(binHome, filepath.Join(root, "build", binaryName)) {
			t.Error("isManaged() = true, want false")
		}
	})
}

func TestReplaceDir(t *testing.T) {
	t.Run("moves into place", func(t *testing.T) {
		root := t.TempDir()
		src := filepath.Join(root, "staging")
		writeFile(t, filepath.Join(src, binaryName), "")
		target := filepath.Join(root, "lib", "yuman-v0.1.0-202609181503-x86_64")

		if err := replaceDir(src, target); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(target, binaryName)); err != nil {
			t.Errorf("target should hold the moved package: %v", err)
		}
	})

	t.Run("replaces an existing target", func(t *testing.T) {
		root := t.TempDir()
		src := filepath.Join(root, "staging")
		writeFile(t, filepath.Join(src, binaryName), "")
		target := filepath.Join(root, "lib", "yuman-v0.1.0-202609181503-x86_64")
		writeFile(t, filepath.Join(target, "stale"), "")

		if err := replaceDir(src, target); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(target, binaryName)); err != nil {
			t.Errorf("target should hold the moved package: %v", err)
		}
		if _, err := os.Stat(filepath.Join(target, "stale")); !os.IsNotExist(err) {
			t.Errorf("stale entry should be gone, stat err = %v", err)
		}
	})
}

func TestLinkBinary(t *testing.T) {
	const name = "yuman-v0.1.0-202609181503-x86_64"

	t.Run("creates a relative link", func(t *testing.T) {
		root := t.TempDir()
		binHome := filepath.Join(root, ".local", "bin")
		if err := linkBinary(binHome, name); err != nil {
			t.Fatal(err)
		}

		link := filepath.Join(binHome, binaryName)
		target, err := os.Readlink(link)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join("..", "lib", name, binaryName)
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
		old := "yuman-v0.0.9-202609181400-x86_64"
		if err := os.Symlink(filepath.Join("..", "lib", old, binaryName), filepath.Join(binHome, binaryName)); err != nil {
			t.Fatal(err)
		}

		if err := linkBinary(binHome, name); err != nil {
			t.Fatal(err)
		}
		target, err := os.Readlink(filepath.Join(binHome, binaryName))
		if err != nil {
			t.Fatal(err)
		}
		if want := filepath.Join("..", "lib", name, binaryName); target != want {
			t.Errorf("symlink target = %q, want %q", target, want)
		}
		if _, err := os.Lstat(filepath.Join(binHome, binaryName+".tmp")); !os.IsNotExist(err) {
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
		version, timestamp, arch, ok := parseName(name)
		if !ok {
			t.Fatalf("test package name %q does not parse", name)
		}
		return Package{Name: name, Path: path, Version: version, Timestamp: timestamp, Arch: arch}
	}

	t.Run("first install keeps everything", func(t *testing.T) {
		libHome := t.TempDir()
		other := newPackage(t, libHome, "yuman-v0.0.9-202609181400-x86_64")

		if err := removeReplacedPackages(Package{}, false, []Package{other}, "", true); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(other.Path); err != nil {
			t.Errorf("package %q should be kept on a first install: %v", other.Name, err)
		}
	})

	t.Run("single package removed", func(t *testing.T) {
		libHome := t.TempDir()
		old := newPackage(t, libHome, "yuman-v0.0.9-202609181400-x86_64")

		if err := removeReplacedPackages(old, true, []Package{old}, "yuman-v0.1.0-202609181503-x86_64", true); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(old.Path); !os.IsNotExist(err) {
			t.Errorf("old package should be removed, stat err = %v", err)
		}
	})

	t.Run("several packages kept", func(t *testing.T) {
		libHome := t.TempDir()
		first := newPackage(t, libHome, "yuman-v0.0.8-202609181300-x86_64")
		second := newPackage(t, libHome, "yuman-v0.0.9-202609181400-x86_64")

		if err := removeReplacedPackages(second, true, []Package{second, first}, "", false); err != nil {
			t.Fatal(err)
		}
		for _, pkg := range []Package{first, second} {
			if _, err := os.Stat(pkg.Path); err != nil {
				t.Errorf("package %q should be kept: %v", pkg.Name, err)
			}
		}
	})
}

// writeTarball builds a yuman-<version>-<timestamp>-<arch>.tar.gz holding the
// same root entries as a real release, with the binary at the given mode.
func writeTarball(t *testing.T, dir string, binaryMode int64, entries ...string) string {
	t.Helper()

	path := filepath.Join(dir, "yuman-v0.1.0-202609181503-x86_64.tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	contents := map[string]string{
		binaryName:  "#!/bin/sh\necho yuman\n",
		"README.md": "readme",
		"LICENSE":   "license",
	}
	for _, name := range entries {
		mode := int64(0o644)
		if name == binaryName {
			mode = binaryMode
		}
		content := contents[name]
		hdr := &tar.Header{Name: name, Mode: mode, Size: int64(len(content)), Typeflag: tar.TypeReg}
		if err = tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err = tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
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

func TestInstallPackage(t *testing.T) {
	const name = "yuman-v0.1.0-202609181503-x86_64"
	latest := Release{
		Version:   "v0.1.0",
		Timestamp: "202609181503",
		Arch:      "x86_64",
		FileName:  name + ".tar.gz",
	}

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
		binary := filepath.Join(dirs.libHome, name, binaryName)
		info, err := os.Stat(binary)
		if err != nil {
			t.Fatalf("installed binary is missing: %v", err)
		}
		if got := info.Mode().Perm(); got != 0o755 {
			t.Errorf("binary mode = %v, want 0755", got)
		}
		if _, err = os.Stat(filepath.Join(dirs.libHome, name, "README.md")); err != nil {
			t.Errorf("package files were not extracted: %v", err)
		}

		link := filepath.Join(dirs.binHome, binaryName)
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
			if strings.HasPrefix(entry.Name(), ".yuman-install-") {
				t.Errorf("staging directory %q was left behind", entry.Name())
			}
		}
	}

	t.Run("fresh install", func(t *testing.T) {
		dirs := newLayout(t)
		// A release binary that lost its executable bit must be fixed up.
		pkg := writeTarball(t, t.TempDir(), 0o644, binaryName, "README.md", "LICENSE")

		if err := installPackage(context.Background(), dirs, latest, pkg, Package{}, false, nil, false); err != nil {
			t.Fatal(err)
		}
		assertInstalled(t, dirs)
	})

	t.Run("update removes the replaced package", func(t *testing.T) {
		dirs := newLayout(t)
		old := Package{
			Name:      "yuman-v0.0.9-202609181400-x86_64",
			Path:      filepath.Join(dirs.libHome, "yuman-v0.0.9-202609181400-x86_64"),
			Version:   "v0.0.9",
			Timestamp: "202609181400",
			Arch:      "x86_64",
		}
		if err := os.MkdirAll(old.Path, 0o755); err != nil {
			t.Fatal(err)
		}
		pkg := writeTarball(t, t.TempDir(), 0o755, binaryName, "README.md")

		if err := installPackage(context.Background(), dirs, latest, pkg, old, true, []Package{old}, false); err != nil {
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
		for _, build := range []string{"202609181300", "202609181400"} {
			pkg := Package{
				Name:      "yuman-v0.0.9-" + build + "-x86_64",
				Path:      filepath.Join(dirs.libHome, "yuman-v0.0.9-"+build+"-x86_64"),
				Version:   "v0.0.9",
				Timestamp: build,
				Arch:      "x86_64",
			}
			if err := os.MkdirAll(pkg.Path, 0o755); err != nil {
				t.Fatal(err)
			}
			old = append(old, pkg)
		}
		pkg := writeTarball(t, t.TempDir(), 0o755, binaryName, "README.md")

		if err := installPackage(context.Background(), dirs, latest, pkg, old[0], true, old, false); err != nil {
			t.Fatal(err)
		}
		assertInstalled(t, dirs)
		for _, replaced := range old {
			if _, err := os.Stat(replaced.Path); err != nil {
				t.Errorf("package %q should be kept: %v", replaced.Name, err)
			}
		}
	})

	t.Run("rejects an archive without the binary", func(t *testing.T) {
		dirs := newLayout(t)
		pkg := writeTarball(t, t.TempDir(), 0o755, "README.md", "LICENSE")

		if err := installPackage(context.Background(), dirs, latest, pkg, Package{}, false, nil, false); err == nil {
			t.Error("installPackage() = nil error, want layout error")
		}
	})
}
