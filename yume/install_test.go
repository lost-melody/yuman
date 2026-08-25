package yume

import (
	"path/filepath"
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
