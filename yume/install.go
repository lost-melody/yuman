package yume

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// yumeLocaleDomain is the gettext domain used for yume's translations.
const yumeLocaleDomain = "fcitx5-yume"

// environmentFileName is the systemd environment.d file that tells fcitx5
// where to find user-installed addons.
const environmentFileName = "60-yume.conf"

// installLayout holds the destination directories for a yume install.
type installLayout struct {
	libDir    string // parent of fcitx5/, e.g. /usr/lib/x86_64-linux-gnu
	shareDir  string // base share dir, e.g. /usr/share or ~/.local/share
	yumeDir   string // yume data dir, ~/.local/share/yume or /usr/share/yume
	localeDir string // gettext locale dir
	binDir    string // executables dir, /usr/bin or ~/.local/bin
	fontsDir  string // bundled fonts dir, /usr/share/fonts/yume or ~/.local/share/fonts/yume
}

// installOp is a single file copy to perform during installation.
type installOp struct {
	src  string
	dest string
}

var (
	MsgInstallUserDirs = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstallUserDirs",
			Other: "user directories",
		},
	}
	MsgInstallSystemDirs = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstallSystemDirs",
			Other: "system directories",
		},
	}
	MsgErrResolveLayout = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrResolveLayout",
			Other: "resolve install layout",
		},
	}
	MsgErrCreateTempDir = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrCreateTempDir",
			Other: "create temp dir",
		},
	}
	MsgErrSystemInstallPkexec = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrSystemInstallPkexec",
			Other: "system install requires pkexec",
		},
	}
	MsgErrInstallFailed = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrInstallFailed",
			Other: "install failed",
		},
	}
	MsgErrCreateEnvironmentDir = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrCreateEnvironmentDir",
			Other: "create environment.d directory",
		},
	}
	MsgWarnLocaleSkipped = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "WarnLocaleSkipped",
			Other: "locale files were not installed (pkexec was declined)",
		},
	}

	MsgInstallingPackage = func(pkg, target string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "InstallingPackage",
				Other: "Installing package '{{.Package}}' into {{.Target}}...",
			},
			TemplateData: map[string]any{
				"Package": pkg,
				"Target":  target,
			},
		}
	}
	MsgDownloadingPackage = func(url string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "DownloadingPackage",
				Other: "Downloading package '{{.URL}}'...",
			},
			TemplateData: map[string]any{
				"URL": url,
			},
		}
	}
	MsgErrDownloadPackage = func(url string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrDownloadPackage",
				Other: "download package '{{.URL}}'",
			},
			TemplateData: map[string]any{
				"URL": url,
			},
		}
	}
	MsgErrArchMismatch = func(pkgArch, hostArch string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrArchMismatch",
				Other: "package architecture '{{.PackageArch}}' does not match this machine ({{.HostArch}})",
			},
			TemplateData: map[string]any{
				"PackageArch": pkgArch,
				"HostArch":    hostArch,
			},
		}
	}
	MsgInstallingFile = func(src, dest string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "InstallingFile",
				Other: "{{.Source}} -> {{.Destination}}",
			},
			TemplateData: map[string]any{
				"Source":      src,
				"Destination": dest,
			},
		}
	}
	MsgErrStatPackage = func(pkg string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrStatPackage",
				Other: "stat package '{{.Package}}'",
			},
			TemplateData: map[string]any{"Package": pkg},
		}
	}
	MsgErrUnsupportedPackage = func(pkg string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrUnsupportedPackage",
				Other: "unsupported package '{{.Package}}': expected a directory or .tar.gz archive",
			},
			TemplateData: map[string]any{"Package": pkg},
		}
	}
	MsgErrOpenGzip = func(pkg string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrOpenGzip",
				Other: "open gzip stream from '{{.Package}}'",
			},
			TemplateData: map[string]any{"Package": pkg},
		}
	}
	MsgErrArchiveAbsolutePath = func(name string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrArchiveAbsolutePath",
				Other: "archive entry has absolute path: '{{.Name}}'",
			},
			TemplateData: map[string]any{"Name": name},
		}
	}
	MsgErrArchiveEscape = func(name string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrArchiveEscape",
				Other: "archive entry escapes destination: '{{.Name}}'",
			},
			TemplateData: map[string]any{"Name": name},
		}
	}
	MsgErrNoInstallableFiles = func(dir string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrNoInstallableFiles",
				Other: "no installable files found in '{{.Directory}}'",
			},
			TemplateData: map[string]any{"Directory": dir},
		}
	}
	MsgErrWriteEnvironmentFile = func(path string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrWriteEnvironmentFile",
				Other: "write '{{.Path}}'",
			},
			TemplateData: map[string]any{"Path": path},
		}
	}
)

// wrapError localizes cfg and wraps cause, preserving the error chain.
func wrapError(cfg *i18n.LocalizeConfig, cause error) error {
	return fmt.Errorf("%s: %w", tr.Localize(cfg), cause)
}

// InstallYume installs yume into user directories or system directories.
//
//   - pkgPath should be a directory, or a '.tar.gz' file.
//
// Files in the package directory or the archive:
//
//   - lib/fcitx5/libyume.so
//   - share/fcitx5/addon/yume.conf
//   - share/fcitx5/inputmethod/yume.conf
//   - share/yume/*
//   - locale/fcitx5-yume-<LANG>.mo
//
// If installing into user directories, use '~/.local/lib/`, '~/.local/share/`;
// or if installing into system directories, use '/usr/lib/`, `/usr/share/`.
// yume always runs as a non-root user: system installs run a single shell
// command through pkexec, while user installs write straight into the home
// directory. 'yume' data is installed into '~/.local/share/yume/' for user
// installs or '/usr/share/yume/' for system installs, and 'libyume.so' may
// need to be installed into '/usr/lib/x86_64-linux-gnu/fcitx5/', where the
// 'libclassicui.so' is located.
//
// Locale files are always installed into the system locale directory because
// fcitx5 does not load translations from user directories. For user installs
// they are installed through pkexec as a best-effort step: a declined pkexec
// prompt only skips the locale files and does not abort the installation.
func InstallYume(ctx context.Context, pkgPath string, userDirs bool, verbose bool) (err error) {
	displayPath := pkgPath
	if isURL(pkgPath) {
		fmt.Println(tr.Localize(MsgDownloadingPackage(pkgPath)))
		var downloaded string
		downloaded, err = downloadPackage(ctx, pkgPath)
		if err != nil {
			return wrapError(MsgErrDownloadPackage(pkgPath), err)
		}
		defer func() {
			_ = os.Remove(downloaded)
		}()
		pkgPath = downloaded
	}

	target := tr.Localize(&MsgInstallUserDirs)
	if !userDirs {
		target = tr.Localize(&MsgInstallSystemDirs)
	}
	fmt.Println(tr.Localize(MsgInstallingPackage(displayPath, target)))

	layout, err := resolveLayout(userDirs)
	if err != nil {
		return wrapError(&MsgErrResolveLayout, err)
	}

	info, err := os.Stat(pkgPath)
	if err != nil {
		return wrapError(MsgErrStatPackage(pkgPath), err)
	}
	if !info.IsDir() && !isTarGz(pkgPath) {
		return tr.LocalizeError(MsgErrUnsupportedPackage(pkgPath))
	}

	srcDir := pkgPath
	if !info.IsDir() {
		var tmpDir string
		tmpDir, err = os.MkdirTemp("", "yume-install-*")
		if err != nil {
			return wrapError(&MsgErrCreateTempDir, err)
		}
		defer func() {
			_ = os.RemoveAll(tmpDir)
		}()
		if err = extractTarGz(ctx, pkgPath, tmpDir); err != nil {
			return err
		}
		srcDir = tmpDir
	}

	if err = checkArch(srcDir); err != nil {
		return err
	}

	if userDirs {
		if err = writeFcitx5Environment(); err != nil {
			return err
		}
	}

	if err = installFromDir(ctx, srcDir, layout, userDirs, verbose); err != nil {
		return err
	}

	printVersionInfo(srcDir)
	return nil
}

// resolveLayout computes the destination directories for an install.
func resolveLayout(userDirs bool) (installLayout, error) {
	layout := installLayout{}
	if userDirs {
		home, err := os.UserHomeDir()
		if err != nil {
			return installLayout{}, err
		}
		dataHome, err := userDataHome()
		if err != nil {
			return installLayout{}, err
		}
		layout.libDir = filepath.Join(home, ".local", "lib")
		layout.shareDir = dataHome
		layout.yumeDir = filepath.Join(dataHome, "yume")
		layout.binDir = filepath.Join(home, ".local", "bin")
		layout.fontsDir = filepath.Join(dataHome, "fonts", "yume")
	} else {
		layout.libDir = findSystemLibParent()
		layout.shareDir = "/usr/share"
		layout.yumeDir = systemYumeDir
		layout.binDir = "/usr/bin"
		layout.fontsDir = "/usr/share/fonts/yume"
	}
	// Locale files always go to the system directory: fcitx5 does not load
	// translations from user directories.
	layout.localeDir = "/usr/share/locale"
	return layout, nil
}

// InstalledTargets reports whether yume is installed in the user and/or
// system directories, based on the presence of the yume addon library.
func InstalledTargets() (user, system bool) {
	if home, err := os.UserHomeDir(); err == nil {
		userLib := filepath.Join(home, ".local", "lib", "fcitx5", libYumeName)
		if _, err := os.Stat(userLib); err == nil {
			user = true
		}
	}
	system = len(globLibYumeFiles()) > 0
	return user, system
}

// DetectInstallTarget reports where an existing yume installation lives,
// preferring the user directories over the system directories. found is false
// when neither location contains the yume addon library.
func DetectInstallTarget() (user bool, found bool) {
	user, system := InstalledTargets()
	if user {
		return true, true
	}
	if system {
		return false, true
	}
	return false, false
}

// userDataHome returns the user data base directory, honoring XDG_DATA_HOME
// and falling back to ~/.local/share when it is unset or empty.
func userDataHome() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}

// findSystemLibParent returns the directory that contains the fcitx5 addon
// libraries, preferring the multiarch directory where libclassicui.so lives.
func findSystemLibParent() string {
	for _, pattern := range []string{
		"/usr/lib/*/fcitx5/libclassicui.so",
		"/usr/lib/fcitx5/libclassicui.so",
	} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		if len(matches) > 0 {
			// libclassicui.so lives at <libParent>/fcitx5/libclassicui.so.
			return filepath.Dir(filepath.Dir(matches[0]))
		}
	}
	return "/usr/lib"
}

// fcitx5AddonDirs returns the fcitx5 addon directories, starting with the user
// directory and followed by the system directories where libclassicui.so lives.
func fcitx5AddonDirs() []string {
	dirs := []string{"${HOME}/.local/lib/fcitx5", "/usr/lib/fcitx5"}
	parent := findSystemLibParent()
	if parent != "/usr/lib" {
		dirs = append(dirs, filepath.Join(parent, "fcitx5"))
	}
	return dirs
}

// writeFcitx5Environment writes the FCITX_ADDON_DIRS environment variable so
// fcitx5 discovers the addon installed into the user directories.
func writeFcitx5Environment() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ".config", "environment.d")
	path := filepath.Join(configDir, environmentFileName)
	content := "FCITX_ADDON_DIRS=" + strings.Join(fcitx5AddonDirs(), ":") + "\n"

	if err = os.MkdirAll(configDir, 0o755); err != nil {
		return wrapError(&MsgErrCreateEnvironmentDir, err)
	}
	if err = os.WriteFile(path, []byte(content), 0o644); err != nil {
		return wrapError(MsgErrWriteEnvironmentFile(path), err)
	}
	return nil
}

// destinationFor maps a package-relative path to its install destination. The
// second return value reports whether the entry should be installed at all.
func destinationFor(rel string, layout installLayout) (string, bool) {
	rel = filepath.ToSlash(filepath.Clean(rel))
	switch {
	case strings.HasPrefix(rel, "lib/fcitx5/"):
		sub := strings.TrimPrefix(rel, "lib/")
		return filepath.Join(layout.libDir, filepath.FromSlash(sub)), true
	case strings.HasPrefix(rel, "share/fcitx5/"):
		sub := strings.TrimPrefix(rel, "share/")
		return filepath.Join(layout.shareDir, filepath.FromSlash(sub)), true
	case strings.HasPrefix(rel, "share/yume/Fonts/"):
		sub := strings.TrimPrefix(rel, "share/yume/Fonts/")
		if !strings.EqualFold(filepath.Ext(sub), ".ttf") {
			return "", false
		}
		// Fonts are installed flat into the font directory.
		return filepath.Join(layout.fontsDir, filepath.Base(filepath.FromSlash(sub))), true
	case strings.HasPrefix(rel, "share/yume/"):
		sub := strings.TrimPrefix(rel, "share/yume/")
		return filepath.Join(layout.yumeDir, filepath.FromSlash(sub)), true
	case strings.HasPrefix(rel, "bin/"):
		sub := strings.TrimPrefix(rel, "bin/")
		return filepath.Join(layout.binDir, filepath.FromSlash(sub)), true
	case strings.HasPrefix(rel, "locale/"):
		base := filepath.Base(rel)
		if strings.HasPrefix(base, yumeLocaleDomain+"-") && strings.HasSuffix(base, ".mo") {
			lang := strings.TrimSuffix(strings.TrimPrefix(base, yumeLocaleDomain+"-"), ".mo")
			return filepath.Join(layout.localeDir, lang, "LC_MESSAGES", yumeLocaleDomain+".mo"), true
		}
		return "", false
	default:
		return "", false
	}
}

func isTarGz(pkgPath string) bool {
	lower := strings.ToLower(pkgPath)
	return strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz")
}

// isURL reports whether pkgPath is an http(s) URL that must be downloaded.
func isURL(pkgPath string) bool {
	return strings.HasPrefix(pkgPath, "http://") || strings.HasPrefix(pkgPath, "https://")
}

// downloadPackage downloads the package at url into a temporary .tar.gz file
// and returns its path. The caller is responsible for removing the file.
func downloadPackage(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %s", resp.Status)
	}

	f, err := os.CreateTemp("", "yume-download-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()
	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

// printVersionInfo prints the version, build and arch fields from the package's
// share/yume/VERSION file, ignoring everything else.
func printVersionInfo(srcDir string) {
	v, err := PackageVersion(srcDir)
	if err != nil {
		return
	}
	fmt.Printf("%+v\n", v)
}

// hostArch is the local machine architecture, expressed as a runtime.GOARCH
// value. It is a variable so tests can override it.
var hostArch = runtime.GOARCH

// machineArch maps a Go architecture name (runtime.GOARCH) to the uname -m
// style used by share/yume/VERSION.
func machineArch(goarch string) string {
	switch goarch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	case "386":
		return "i686"
	case "loong64":
		return "loongarch64"
	default:
		return goarch
	}
}

// checkArch verifies that the arch field in the package's share/yume/VERSION
// matches the local machine. A missing or unreadable VERSION (or arch field) is
// tolerated; only an explicit mismatch is an error.
func checkArch(srcDir string) error {
	v, err := PackageVersion(srcDir)
	if err != nil || v.Arch == "" {
		return nil
	}
	local := machineArch(hostArch)
	if v.Arch != local {
		return tr.LocalizeError(MsgErrArchMismatch(v.Arch, local))
	}
	return nil
}

// extractTarGz unpacks a gzipped tarball into destDir, preserving file modes
// and rejecting entries that escape destDir.
func extractTarGz(ctx context.Context, pkgPath, destDir string) error {
	f, err := os.Open(pkgPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return wrapError(MsgErrOpenGzip(pkgPath), err)
	}
	defer func() {
		_ = gz.Close()
	}()

	tarReader := tar.NewReader(gz)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		hdr, err := tarReader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		clean := filepath.Clean(hdr.Name)
		if filepath.IsAbs(clean) {
			return tr.LocalizeError(MsgErrArchiveAbsolutePath(hdr.Name))
		}
		target := filepath.Join(destDir, clean)
		if !withinDir(destDir, target) {
			return tr.LocalizeError(MsgErrArchiveEscape(hdr.Name))
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode).Perm()); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(hdr.Mode).Perm())
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tarReader); err != nil {
				_ = out.Close()
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
		}
	}
}

// withinDir reports whether target is destDir or a path underneath it.
func withinDir(destDir, target string) bool {
	rel, err := filepath.Rel(destDir, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

// installFromDir walks an unpacked package directory and installs every
// recognized file. All copies are assembled into a single shell command that is
// run once, optionally under pkexec for system installs.
// installFromDir walks an unpacked package directory and installs every
// recognized file. All copies are assembled into a single shell command that is
// run once. User installs write non-locale files directly and try to install
// locale files through pkexec; system installs run everything through pkexec.
func installFromDir(ctx context.Context, srcDir string, layout installLayout, userDirs bool, verbose bool) error {
	var normalOps, localeOps []installOp
	normalDirs := map[string]bool{}
	localeDirs := map[string]bool{}
	var normalDirList, localeDirList []string

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		dest, ok := destinationFor(rel, layout)
		if !ok {
			return nil
		}

		op := installOp{src: path, dest: dest}
		if strings.HasPrefix(filepath.ToSlash(rel), "locale/") {
			localeOps = append(localeOps, op)
			if parent := filepath.Dir(dest); !localeDirs[parent] {
				localeDirs[parent] = true
				localeDirList = append(localeDirList, parent)
			}
		} else {
			normalOps = append(normalOps, op)
			if parent := filepath.Dir(dest); !normalDirs[parent] {
				normalDirs[parent] = true
				normalDirList = append(normalDirList, parent)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(normalOps) == 0 && len(localeOps) == 0 {
		return tr.LocalizeError(MsgErrNoInstallableFiles(srcDir))
	}

	if verbose {
		all := append(append([]installOp{}, normalOps...), localeOps...)
		for _, op := range all {
			fmt.Printf("  %s\n", tr.Localize(MsgInstallingFile(op.src, op.dest)))
		}
	}

	if userDirs {
		if len(normalOps) > 0 {
			script := buildInstallScript(normalDirList, normalOps)
			if err := runShellScript(ctx, script, false, &MsgErrSystemInstallPkexec, &MsgErrInstallFailed); err != nil {
				return err
			}
		}
		if len(localeOps) > 0 {
			script := buildInstallScript(localeDirList, localeOps)
			if err := runShellScript(ctx, script, true, &MsgErrSystemInstallPkexec, &MsgErrInstallFailed); err != nil {
				// Locale files are best-effort during user installs; a declined
				// pkexec prompt must not abort the installation.
				fmt.Printf("%s\n", tr.Localize(&MsgWarnLocaleSkipped))
			}
		}
		return nil
	}

	allOps := append(append([]installOp{}, normalOps...), localeOps...)
	allDirs := append(append([]string{}, normalDirList...), localeDirList...)
	script := buildInstallScript(allDirs, allOps)
	return runShellScript(ctx, script, true, &MsgErrSystemInstallPkexec, &MsgErrInstallFailed)
}

// buildInstallScript assembles the copy operations into a single "&&"-joined
// shell command.
func buildInstallScript(dirs []string, ops []installOp) string {
	var parts []string
	if len(dirs) > 0 {
		quoted := make([]string, len(dirs))
		for i, d := range dirs {
			quoted[i] = shellQuote(d)
		}
		parts = append(parts, "mkdir -p "+strings.Join(quoted, " "))
	}
	for _, op := range ops {
		parts = append(parts, "cp "+shellQuote(op.src)+" "+shellQuote(op.dest))
	}
	return strings.Join(parts, " && ")
}
