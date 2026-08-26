package yume

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUserDataHome(t *testing.T) {
	t.Run("XDG_DATA_HOME set", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "/custom/data")
		got, err := userDataHome()
		if err != nil {
			t.Fatal(err)
		}
		if want := "/custom/data"; got != want {
			t.Errorf("userDataHome() = %q, want %q", got, want)
		}
	})

	t.Run("XDG_DATA_HOME fallback", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "")
		t.Setenv("HOME", "/home/test")
		got, err := userDataHome()
		if err != nil {
			t.Fatal(err)
		}
		if want := filepath.Join("/home/test", ".local", "share"); got != want {
			t.Errorf("userDataHome() = %q, want %q", got, want)
		}
	})
}

func TestResolveLayoutUser(t *testing.T) {
	t.Setenv("HOME", "/home/test")
	t.Setenv("XDG_DATA_HOME", "/custom/data")

	layout, err := resolveLayout(true)
	if err != nil {
		t.Fatal(err)
	}
	if want := "/custom/data"; layout.shareDir != want {
		t.Errorf("shareDir = %q, want %q", layout.shareDir, want)
	}
	if want := filepath.Join("/custom/data", "yume"); layout.yumeDir != want {
		t.Errorf("yumeDir = %q, want %q", layout.yumeDir, want)
	}
	if want := filepath.Join("/home/test", ".local", "lib"); layout.libDir != want {
		t.Errorf("libDir = %q, want %q", layout.libDir, want)
	}
}

func TestResolveLayoutUserFallback(t *testing.T) {
	t.Setenv("HOME", "/home/test")
	t.Setenv("XDG_DATA_HOME", "")

	layout, err := resolveLayout(true)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/home/test", ".local", "share")
	if layout.shareDir != want {
		t.Errorf("shareDir = %q, want %q", layout.shareDir, want)
	}
	if want := filepath.Join(want, "yume"); layout.yumeDir != want {
		t.Errorf("yumeDir = %q, want %q", layout.yumeDir, want)
	}
}

func TestResolveLayoutSystem(t *testing.T) {
	t.Setenv("HOME", "/home/test")
	t.Setenv("XDG_DATA_HOME", "/custom/data")

	layout, err := resolveLayout(false)
	if err != nil {
		t.Fatal(err)
	}
	if want := "/usr/share"; layout.shareDir != want {
		t.Errorf("shareDir = %q, want %q", layout.shareDir, want)
	}
	if want := "/usr/share/yume"; layout.yumeDir != want {
		t.Errorf("yumeDir = %q, want %q", layout.yumeDir, want)
	}
	if want := "/usr/share/locale"; layout.localeDir != want {
		t.Errorf("localeDir = %q, want %q", layout.localeDir, want)
	}
}

func TestResolveLayoutBinAndFontsDirs(t *testing.T) {
	t.Run("user", func(t *testing.T) {
		t.Setenv("HOME", "/home/test")
		t.Setenv("XDG_DATA_HOME", "")
		layout, err := resolveLayout(true)
		if err != nil {
			t.Fatal(err)
		}
		if want := filepath.Join("/home/test", ".local", "bin"); layout.binDir != want {
			t.Errorf("binDir = %q, want %q", layout.binDir, want)
		}
		if want := filepath.Join("/home/test", ".local", "share", "fonts", "yume"); layout.fontsDir != want {
			t.Errorf("fontsDir = %q, want %q", layout.fontsDir, want)
		}
	})

	t.Run("system", func(t *testing.T) {
		layout, err := resolveLayout(false)
		if err != nil {
			t.Fatal(err)
		}
		if want := "/usr/bin"; layout.binDir != want {
			t.Errorf("binDir = %q, want %q", layout.binDir, want)
		}
		if want := "/usr/share/fonts/yume"; layout.fontsDir != want {
			t.Errorf("fontsDir = %q, want %q", layout.fontsDir, want)
		}
	})
}

func TestDestinationForBinAndFonts(t *testing.T) {
	layout := installLayout{
		binDir:   "/bin",
		fontsDir: "/share/fonts/yume",
		yumeDir:  "/share/yume",
	}
	tests := []struct {
		rel     string
		want    string
		install bool
	}{
		{"bin/yume-compile", "/bin/yume-compile", true},
		{"share/yume/Fonts/yume.ttf", "/share/fonts/yume/yume.ttf", true},
		{"share/yume/Fonts/sub/yume.ttf", "/share/fonts/yume/yume.ttf", true},
		{"share/yume/Fonts/readme.txt", "", false},
		{"share/yume/conf.yaml", "/share/yume/conf.yaml", true},
	}
	for _, tt := range tests {
		t.Run(tt.rel, func(t *testing.T) {
			got, ok := destinationFor(tt.rel, layout)
			if ok != tt.install || got != tt.want {
				t.Errorf("destinationFor(%q) = (%q, %v), want (%q, %v)", tt.rel, got, ok, tt.want, tt.install)
			}
		})
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"https://example.com/yume.tar.gz", true},
		{"http://example.com/yume.tar.gz", true},
		{"ftp://example.com/yume.tar.gz", false},
		{"/local/path/yume.tar.gz", false},
	}
	for _, tt := range tests {
		if got := isURL(tt.in); got != tt.want {
			t.Errorf("isURL(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestDownloadPackage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("faketarball"))
	}))
	defer server.Close()

	path, err := downloadPackage(context.Background(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(path)
	}()

	if !strings.HasSuffix(path, ".tar.gz") {
		t.Errorf("downloadPackage() path = %q, want .tar.gz suffix", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "faketarball" {
		t.Errorf("downloadPackage() content = %q, want %q", data, "faketarball")
	}
}

func TestMachineArch(t *testing.T) {
	tests := []struct {
		goarch string
		want   string
	}{
		{"amd64", "x86_64"},
		{"arm64", "aarch64"},
		{"386", "i686"},
		{"loong64", "loongarch64"},
		{"riscv64", "riscv64"},
	}
	for _, tt := range tests {
		if got := machineArch(tt.goarch); got != tt.want {
			t.Errorf("machineArch(%q) = %q, want %q", tt.goarch, got, tt.want)
		}
	}
}

func TestCheckArch(t *testing.T) {
	old := hostArch
	hostArch = "amd64" // machineArch -> x86_64
	t.Cleanup(func() { hostArch = old })

	writeVersion := func(t *testing.T, content string) string {
		t.Helper()
		srcDir := t.TempDir()
		versionFile := filepath.Join(srcDir, "share", "yume", "VERSION")
		if err := os.MkdirAll(filepath.Dir(versionFile), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(versionFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return srcDir
	}

	t.Run("match", func(t *testing.T) {
		if err := checkArch(writeVersion(t, "arch=x86_64\n")); err != nil {
			t.Errorf("checkArch() = %v, want nil", err)
		}
	})
	t.Run("mismatch", func(t *testing.T) {
		if err := checkArch(writeVersion(t, "arch=aarch64\n")); err == nil {
			t.Error("checkArch() = nil, want error")
		}
	})
	t.Run("missing arch line", func(t *testing.T) {
		if err := checkArch(writeVersion(t, "version=1.0\n")); err != nil {
			t.Errorf("checkArch() = %v, want nil", err)
		}
	})
	t.Run("missing version file", func(t *testing.T) {
		if err := checkArch(t.TempDir()); err != nil {
			t.Errorf("checkArch() = %v, want nil", err)
		}
	})
}

func TestInstallYumeRejectsArchMismatch(t *testing.T) {
	old := hostArch
	hostArch = "amd64" // machineArch -> x86_64
	t.Cleanup(func() { hostArch = old })

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	srcDir := t.TempDir()
	lib := filepath.Join(srcDir, "lib", "fcitx5", libYumeName)
	if err := os.MkdirAll(filepath.Dir(lib), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lib, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	versionFile := filepath.Join(srcDir, "share", "yume", "VERSION")
	if err := os.MkdirAll(filepath.Dir(versionFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(versionFile, []byte("arch=aarch64\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := InstallYume(context.Background(), srcDir, true, false); err == nil {
		t.Fatal("InstallYume() = nil error, want arch mismatch")
	}

	// The arch check must block the install before anything is written.
	if _, err := os.Stat(filepath.Join(home, ".local", "lib", "fcitx5", libYumeName)); !os.IsNotExist(err) {
		t.Errorf("libyume was installed despite arch mismatch: %v", err)
	}
}

func TestInstalledTargetsUser(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	user, system := InstalledTargets()
	if user || system {
		t.Fatalf("InstalledTargets() = (%v, %v), want (false, false)", user, system)
	}

	lib := filepath.Join(home, ".local", "lib", "fcitx5", libYumeName)
	if err := os.MkdirAll(filepath.Dir(lib), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lib, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	user, system = InstalledTargets()
	if !user || system {
		t.Fatalf("InstalledTargets() = (%v, %v), want (true, false)", user, system)
	}
}
