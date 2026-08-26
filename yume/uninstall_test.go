package yume

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestBuildRemovalScript(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		dirs     []string
		tidyDirs []string
		want     string
	}{
		{
			name:     "all groups",
			files:    []string{"/a/b"},
			dirs:     []string{"/c/d"},
			tidyDirs: []string{"/e"},
			want:     "rm -f '/a/b' && rm -rf '/c/d' && rmdir --ignore-fail-on-non-empty '/e'",
		},
		{
			name:  "files only",
			files: []string{"/a/b", "/a/c"},
			want:  "rm -f '/a/b' '/a/c'",
		},
		{
			name: "dirs only",
			dirs: []string{"/c/d"},
			want: "rm -rf '/c/d'",
		},
		{
			name:     "tidy dirs only",
			tidyDirs: []string{"/e", "/f"},
			want:     "rmdir --ignore-fail-on-non-empty '/e' '/f'",
		},
		{
			name:  "path with a single quote",
			files: []string{"/a/b'c"},
			want:  "rm -f '/a/b'\\''c'",
		},
		{
			name: "empty plan",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildRemovalScript(tt.files, tt.dirs, tt.tidyDirs); got != tt.want {
				t.Errorf("buildRemovalScript() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExistingPaths(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "exists")
	if err := os.WriteFile(existing, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := existingPaths([]string{existing, filepath.Join(dir, "missing"), dir})
	if want := []string{existing, dir}; !slices.Equal(got, want) {
		t.Errorf("existingPaths() = %v, want %v", got, want)
	}
	if got := existingPaths(nil); got != nil {
		t.Errorf("existingPaths(nil) = %v, want nil", got)
	}
}

func TestYumeDataDir(t *testing.T) {
	if got, want := yumeDataDir("/home/test"), filepath.Join("/home/test", ".local", "share", "yume"); got != want {
		t.Errorf("yumeDataDir() = %q, want %q", got, want)
	}
}

func TestPlanUserUninstall(t *testing.T) {
	home := "/home/test"
	plan := planUserUninstall(home, nil)

	wantFiles := []string{
		filepath.Join(home, ".local", "lib", "fcitx5", libYumeName),
		filepath.Join(home, ".local", "share", "fcitx5", "addon", "yume.conf"),
		filepath.Join(home, ".local", "share", "fcitx5", "inputmethod", "yume.conf"),
		filepath.Join(home, ".config", "environment.d", environmentFileName),
		filepath.Join(home, ".local", "bin", compileBinaryName),
	}
	if !slices.Equal(plan.files, wantFiles) {
		t.Errorf("plan.files = %v, want %v", plan.files, wantFiles)
	}
	wantTidy := []string{
		filepath.Join(home, ".local", "lib", "fcitx5"),
		filepath.Join(home, ".local", "share", "fcitx5", "addon"),
		filepath.Join(home, ".local", "share", "fcitx5", "inputmethod"),
		filepath.Join(home, ".local", "share", "fcitx5"),
		filepath.Join(home, ".config", "environment.d"),
		userFontsDir(home),
		filepath.Join(home, ".local", "share", "fonts"),
	}
	if !slices.Equal(plan.tidyDirs, wantTidy) {
		t.Errorf("plan.tidyDirs = %v, want %v", plan.tidyDirs, wantTidy)
	}
	if len(plan.dirs) != 0 {
		t.Errorf("plan.dirs = %v, want empty", plan.dirs)
	}
}

func TestPlanSystemUninstall(t *testing.T) {
	home := "/home/test"
	libYume := filepath.Join("/usr", "lib", "x86_64-linux-gnu", "fcitx5", libYumeName)
	mo := filepath.Join("/usr", "share", "locale", "zh_CN", "LC_MESSAGES", yumeLocaleDomain+".mo")
	wantFiles := []string{
		libYume,
		filepath.Join(systemShareDir, "fcitx5", "addon", "yume.conf"),
		filepath.Join(systemShareDir, "fcitx5", "inputmethod", "yume.conf"),
		filepath.Join("/usr/bin", compileBinaryName),
		mo,
	}
	wantTidy := []string{
		filepath.Dir(libYume),
		filepath.Join(systemShareDir, "fcitx5", "addon"),
		filepath.Join(systemShareDir, "fcitx5", "inputmethod"),
		filepath.Join(systemShareDir, "fcitx5"),
		filepath.Dir(mo),
		filepath.Dir(filepath.Dir(mo)),
		systemFontsDir,
		filepath.Dir(systemFontsDir),
	}

	for _, purge := range []bool{false, true} {
		plan := planSystemUninstall(home, []string{libYume}, []string{mo}, nil, purge)
		if !slices.Equal(plan.files, wantFiles) {
			t.Errorf("purge=%v: plan.files = %v, want %v", purge, plan.files, wantFiles)
		}
		// The data directory holds user data and is only removed when purging.
		var wantDirs []string
		if purge {
			wantDirs = []string{yumeDataDir(home)}
		}
		if !slices.Equal(plan.dirs, wantDirs) {
			t.Errorf("purge=%v: plan.dirs = %v, want %v", purge, plan.dirs, wantDirs)
		}
		if !slices.Equal(plan.tidyDirs, wantTidy) {
			t.Errorf("purge=%v: plan.tidyDirs = %v, want %v", purge, plan.tidyDirs, wantTidy)
		}
	}
}

func TestLocaleMoFiles(t *testing.T) {
	localeDir := t.TempDir()
	mo := filepath.Join(localeDir, "zh_CN", "LC_MESSAGES", yumeLocaleDomain+".mo")
	for _, path := range []string{
		mo,
		filepath.Join(localeDir, "zh_CN", "LC_MESSAGES", "other-domain.mo"),
		filepath.Join(localeDir, "zh_TW", "LC_MESSAGES", yumeLocaleDomain+".mo"),
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got := localeMoFiles(localeDir)
	want := []string{
		filepath.Join(localeDir, "zh_CN", "LC_MESSAGES", yumeLocaleDomain+".mo"),
		filepath.Join(localeDir, "zh_TW", "LC_MESSAGES", yumeLocaleDomain+".mo"),
	}
	if !slices.Equal(got, want) {
		t.Errorf("localeMoFiles() = %v, want %v", got, want)
	}
}

func TestLocaleTidyDirs(t *testing.T) {
	mo := filepath.Join("/usr", "share", "locale", "zh_CN", "LC_MESSAGES", yumeLocaleDomain+".mo")
	want := []string{
		filepath.Join("/usr", "share", "locale", "zh_CN", "LC_MESSAGES"),
		filepath.Join("/usr", "share", "locale", "zh_CN"),
	}
	if got := localeTidyDirs([]string{mo}); !slices.Equal(got, want) {
		t.Errorf("localeTidyDirs() = %v, want %v", got, want)
	}
}

func TestRunUninstallPlanRemovesFilesAndTidiesEmptyDirs(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "fcitx5", "yume.conf")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	plan := uninstallPlan{
		files:    []string{file},
		tidyDirs: []string{filepath.Dir(file)},
	}
	if err := runUninstallPlan(context.Background(), plan, false, false); err != nil {
		t.Fatalf("runUninstallPlan() error = %v", err)
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Errorf("file still exists: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(file)); !os.IsNotExist(err) {
		t.Errorf("empty parent dir still exists: %v", err)
	}
}

func TestRunUninstallPlanSkipsWhenNothingExists(t *testing.T) {
	dir := t.TempDir()
	tidyDir := filepath.Join(dir, "leftover")
	if err := os.Mkdir(tidyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// No removal target exists, so the plan is skipped entirely and even an
	// existing empty tidy dir must not be touched.
	plan := uninstallPlan{
		files:    []string{filepath.Join(dir, "missing")},
		tidyDirs: []string{tidyDir},
	}
	if err := runUninstallPlan(context.Background(), plan, false, false); err != nil {
		t.Fatalf("runUninstallPlan() error = %v", err)
	}
	if _, err := os.Stat(tidyDir); err != nil {
		t.Errorf("tidy dir was removed: %v", err)
	}
}

func TestRunUninstallPlanKeepsNonEmptyTidyDirs(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "yume.conf")
	other := filepath.Join(dir, "other.conf")
	for _, path := range []string{file, other} {
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	plan := uninstallPlan{
		files:    []string{file},
		tidyDirs: []string{dir},
	}
	if err := runUninstallPlan(context.Background(), plan, false, false); err != nil {
		t.Fatalf("runUninstallPlan() error = %v", err)
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Errorf("file still exists: %v", err)
	}
	if _, err := os.Stat(other); err != nil {
		t.Errorf("unrelated file was removed: %v", err)
	}
}

func TestRunUninstallPlanReportsFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "yume.conf")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A read-only parent directory makes rm fail for the file owner.
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o755)
	})

	plan := uninstallPlan{files: []string{file}}
	if err := runUninstallPlan(context.Background(), plan, false, false); err == nil {
		t.Error("runUninstallPlan() = nil error, want failure")
	}
}

// setupUserInstall creates a fake user installation under a temporary HOME
// (including user data inside the yume data directory) and redirects the
// locale glob so tests never touch the real system locale files.
func setupUserInstall(t *testing.T) (home string, files []string) {
	t.Helper()

	home = t.TempDir()
	t.Setenv("HOME", home)

	localeDir := t.TempDir()
	oldLocaleDir := systemLocaleDir
	systemLocaleDir = localeDir
	t.Cleanup(func() {
		systemLocaleDir = oldLocaleDir
	})

	files = planUserUninstall(home, nil).files
	for _, path := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	userData := userDictPath(home)
	if err := os.MkdirAll(filepath.Dir(userData), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(userData, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home, files
}

// userDictPath returns a file inside the yume data directory, standing in
// for user data such as a learned dictionary.
func userDictPath(home string) string {
	return filepath.Join(yumeDataDir(home), "dict", "user.dict")
}

func TestUninstallYumeUserDirsKeepsDataWithoutPurge(t *testing.T) {
	home, files := setupUserInstall(t)

	if err := UninstallYume(context.Background(), true, false, false); err != nil {
		t.Fatalf("UninstallYume() error = %v", err)
	}

	for _, path := range files {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s still exists: %v", path, err)
		}
	}
	// Directories that only contained yume files are tidied up as well.
	for _, path := range planUserUninstall(home, nil).tidyDirs {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("empty dir %s still exists: %v", path, err)
		}
	}
	// User data must survive an uninstall that does not purge.
	if _, err := os.Stat(userDictPath(home)); err != nil {
		t.Errorf("user data was removed: %v", err)
	}
}

func TestUninstallYumeUserDirsPurgesData(t *testing.T) {
	home, files := setupUserInstall(t)

	if err := UninstallYume(context.Background(), true, true, false); err != nil {
		t.Fatalf("UninstallYume() error = %v", err)
	}

	for _, path := range append(slices.Clone(files), yumeDataDir(home)) {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s still exists: %v", path, err)
		}
	}
}

func TestExistingYumeDataDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if dir, ok := ExistingYumeDataDir(); ok {
		t.Errorf("ExistingYumeDataDir() = (%q, true), want missing", dir)
	}

	dataDir := yumeDataDir(home)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if dir, ok := ExistingYumeDataDir(); !ok || dir != dataDir {
		t.Errorf("ExistingYumeDataDir() = (%q, %v), want (%q, true)", dir, ok, dataDir)
	}
}

func TestRunShellScript(t *testing.T) {
	ctx := context.Background()
	if err := runShellScript(ctx, "true", false, nil, nil); err != nil {
		t.Errorf("runShellScript(true) error = %v", err)
	}
	err := runShellScript(ctx, "false", false, &MsgErrSystemInstallPkexec, &MsgErrInstallFailed)
	if err == nil {
		t.Fatal("runShellScript(false) = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "exit status") {
		t.Errorf("runShellScript(false) error = %v, want wrapped exit status", err)
	}
}

func TestShellQuoteAll(t *testing.T) {
	got := shellQuoteAll([]string{"a b", "c'd"})
	want := []string{`'a b'`, `'c'\''d'`}
	if !slices.Equal(got, want) {
		t.Errorf("shellQuoteAll() = %v, want %v", got, want)
	}
}

func TestFontFiles(t *testing.T) {
	fontsDir := t.TempDir()
	ttf := filepath.Join(fontsDir, "yume.ttf")
	upper := filepath.Join(fontsDir, "Yume.TTF")
	other := filepath.Join(fontsDir, "readme.txt")
	for _, path := range []string{ttf, upper, other} {
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got := fontFiles(fontsDir)
	// os.ReadDir returns entries sorted by name, so the uppercase name first.
	want := []string{upper, ttf}
	if !slices.Equal(got, want) {
		t.Errorf("fontFiles() = %v, want %v", got, want)
	}
	if got := fontFiles(filepath.Join(fontsDir, "missing")); got != nil {
		t.Errorf("fontFiles(missing) = %v, want nil", got)
	}
}

func TestFontTidyDirs(t *testing.T) {
	want := []string{"/share/fonts/yume", "/share/fonts"}
	if got := fontTidyDirs("/share/fonts/yume"); !slices.Equal(got, want) {
		t.Errorf("fontTidyDirs() = %v, want %v", got, want)
	}
}

func TestUninstallYumeUserDirsRemovesFonts(t *testing.T) {
	home, _ := setupUserInstall(t)

	fontsDir := userFontsDir(home)
	ttf := filepath.Join(fontsDir, "yume.ttf")
	if err := os.MkdirAll(fontsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ttf, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := UninstallYume(context.Background(), true, false, false); err != nil {
		t.Fatalf("UninstallYume() error = %v", err)
	}

	if _, err := os.Stat(ttf); !os.IsNotExist(err) {
		t.Errorf("font still exists: %v", err)
	}
	if _, err := os.Stat(fontsDir); !os.IsNotExist(err) {
		t.Errorf("font dir still exists: %v", err)
	}
}
