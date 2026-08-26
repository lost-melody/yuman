package yume

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
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

func TestParseVersion(t *testing.T) {
	got := parseVersion("name=yume\nversion=1.2.3\nbuild=abc123\narch=x86_64\nextra=ignored\n")
	want := YumeVersion{Version: "1.2.3", Build: "abc123", Arch: "x86_64"}
	if got != want {
		t.Errorf("parseVersion() = %+v, want %+v", got, want)
	}
}

func TestPackageVersion(t *testing.T) {
	srcDir := t.TempDir()
	writeFile(t, filepath.Join(srcDir, "share", "yume", "VERSION"), "version=1.0\nbuild=abc\narch=x86_64\n")

	got, err := PackageVersion(srcDir)
	if err != nil {
		t.Fatal(err)
	}
	want := YumeVersion{Version: "1.0", Build: "abc", Arch: "x86_64"}
	if got != want {
		t.Errorf("PackageVersion() = %+v, want %+v", got, want)
	}

	if _, err := PackageVersion(t.TempDir()); err == nil {
		t.Error("PackageVersion() on missing VERSION = nil error, want error")
	}
}

func TestVersionLines(t *testing.T) {
	v := YumeVersion{Version: "1.0", Build: "abc", Arch: "x86_64"}
	want := []string{"version=1.0", "build=abc", "arch=x86_64"}
	if got := v.Lines(); !reflect.DeepEqual(got, want) {
		t.Errorf("Lines() = %v, want %v", got, want)
	}
}

func TestInstalledYumeVersion(t *testing.T) {
	oldSystem := systemYumeDir
	t.Cleanup(func() { systemYumeDir = oldSystem })

	userVersion := func(home string) string {
		return filepath.Join(home, "yume", "VERSION")
	}

	t.Run("user preferred", func(t *testing.T) {
		systemYumeDir = t.TempDir()
		home := t.TempDir()
		t.Setenv("XDG_DATA_HOME", home)

		writeFile(t, userVersion(home), "version=user\nbuild=u\narch=x86_64\n")
		writeFile(t, filepath.Join(systemYumeDir, "VERSION"), "version=system\nbuild=s\narch=x86_64\n")

		got, found := InstalledYumeVersion()
		if !found {
			t.Fatal("InstalledYumeVersion() found = false, want true")
		}
		want := YumeVersion{Version: "user", Build: "u", Arch: "x86_64"}
		if got != want {
			t.Errorf("InstalledYumeVersion() = %+v, want %+v", got, want)
		}
	})

	t.Run("system fallback", func(t *testing.T) {
		systemYumeDir = t.TempDir()
		t.Setenv("XDG_DATA_HOME", t.TempDir())

		writeFile(t, filepath.Join(systemYumeDir, "VERSION"), "version=system\nbuild=s\narch=x86_64\n")

		got, found := InstalledYumeVersion()
		if !found {
			t.Fatal("InstalledYumeVersion() found = false, want true")
		}
		want := YumeVersion{Version: "system", Build: "s", Arch: "x86_64"}
		if got != want {
			t.Errorf("InstalledYumeVersion() = %+v, want %+v", got, want)
		}
	})

	t.Run("not installed", func(t *testing.T) {
		systemYumeDir = t.TempDir()
		t.Setenv("XDG_DATA_HOME", t.TempDir())

		got, found := InstalledYumeVersion()
		if found {
			t.Fatalf("InstalledYumeVersion() = (%+v, true), want not found", got)
		}
	})
}
