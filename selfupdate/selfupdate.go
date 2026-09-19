// Package selfupdate upgrades the yuman command-line helper itself from its GitHub
// releases. Every version is installed into the current user's directories
// under $HOME, so root privileges and the XDG environment are never involved.
package selfupdate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/lost-melody/yuman/release"
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// ReleaseRepo is the GitHub repository that hosts yuman releases.
const ReleaseRepo = "lost-melody/yuman"

const (
	// binaryName is the executable named after the package root file.
	binaryName = "yuman"
	// libPrefix is the prefix of a yuman package directory and release asset.
	libPrefix = "yuman-"
	// timestampLen is the length of the UTC timestamp in a release name.
	timestampLen = 12
)

var (
	MsgYumanUpgradeChecking = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumanUpgradeChecking",
			Other: "Checking for the latest yuman release...",
		},
	}
	MsgYumanNotInstalled = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumanNotInstalled",
			Other: "no yuman package is installed",
		},
	}
	MsgErrResolveYumanLayout = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrResolveYumanLayout",
			Other: "resolve yuman install layout",
		},
	}
	MsgErrNoYumanRelease = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrNoYumanRelease",
			Other: "no yuman release found",
		},
	}
	MsgErrCreateYumanTempDir = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrCreateYumanTempDir",
			Other: "create yuman temp dir",
		},
	}

	MsgYumanLayout = func(libHome, binHome string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumanLayout",
				Other: "Install directories: lib '{{.Lib}}', bin '{{.Bin}}'",
			},
			TemplateData: map[string]any{
				"Lib": libHome,
				"Bin": binHome,
			},
		}
	}
	MsgYumanRelease = func(version, timestamp, fileName, size string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumanRelease",
				Other: "Latest yuman release: {{.Version}}-{{.Timestamp}} ({{.FileName}}, {{.Size}})",
			},
			TemplateData: map[string]any{
				"Version":   version,
				"Timestamp": timestamp,
				"FileName":  fileName,
				"Size":      size,
			},
		}
	}
	MsgYumanInstalledPackages = func(packages string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumanInstalledPackages",
				Other: "Installed yuman packages: {{.Packages}}",
			},
			TemplateData: map[string]any{
				"Packages": packages,
			},
		}
	}
	MsgYumanUpToDate = func(version, timestamp string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumanUpToDate",
				Other: "yuman is already up to date ({{.Version}}-{{.Timestamp}})",
			},
			TemplateData: map[string]any{
				"Version":   version,
				"Timestamp": timestamp,
			},
		}
	}
	MsgYumanUpdateAvailable = func(newVersion, newTimestamp, newSize, currentVersion, currentTimestamp string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumanUpdateAvailable",
				Other: "Update available: {{.New}}-{{.NewTimestamp}} (size: {{.NewSize}}) (installed {{.Current}}-{{.CurrentTimestamp}})",
			},
			TemplateData: map[string]any{
				"New":              newVersion,
				"NewTimestamp":     newTimestamp,
				"NewSize":          newSize,
				"Current":          currentVersion,
				"CurrentTimestamp": currentTimestamp,
			},
		}
	}
	MsgYumanWarnNotManaged = func(exe, bin string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumanWarnNotManaged",
				Other: "warning: the running yuman '{{.Path}}' is not the managed command '{{.Bin}}'",
			},
			TemplateData: map[string]any{
				"Path": exe,
				"Bin":  bin,
			},
		}
	}
	MsgYumanInstalling = func(version, timestamp, dir string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumanInstalling",
				Other: "Installing yuman {{.Version}}-{{.Timestamp}} into '{{.Directory}}'...",
			},
			TemplateData: map[string]any{
				"Version":   version,
				"Timestamp": timestamp,
				"Directory": dir,
			},
		}
	}
	MsgYumanExtracting = func(pkg string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumanExtracting",
				Other: "Extracting '{{.Package}}'...",
			},
			TemplateData: map[string]any{
				"Package": pkg,
			},
		}
	}
	MsgYumanLinking = func(link, target string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumanLinking",
				Other: "'{{.Link}}' -> '{{.Target}}'",
			},
			TemplateData: map[string]any{
				"Link":   link,
				"Target": target,
			},
		}
	}
	MsgYumanRemovingPackage = func(path string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumanRemovingPackage",
				Other: "Removing replaced yuman package '{{.Path}}'",
			},
			TemplateData: map[string]any{
				"Path": path,
			},
		}
	}
	MsgYumanWarnMultiplePackages = func(packages string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumanWarnMultiplePackages",
				Other: "multiple yuman packages found ({{.Packages}}); skipping removal of old packages",
			},
			TemplateData: map[string]any{
				"Packages": packages,
			},
		}
	}

	MsgErrFetchYumanRelease = func(url string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrFetchYumanRelease",
				Other: "fetch latest yuman release from '{{.URL}}'",
			},
			TemplateData: map[string]any{
				"URL": url,
			},
		}
	}
	MsgErrYumanNoMatchingAsset = func(arch string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrYumanNoMatchingAsset",
				Other: "no matching linux yuman package for '{{.Arch}}'",
			},
			TemplateData: map[string]any{
				"Arch": arch,
			},
		}
	}
	MsgErrListYumanPackages = func(dir string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrListYumanPackages",
				Other: "list installed yuman packages in '{{.Directory}}'",
			},
			TemplateData: map[string]any{
				"Directory": dir,
			},
		}
	}
	MsgErrCreateYumanLibDir = func(dir string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrCreateYumanLibDir",
				Other: "create yuman lib directory '{{.Directory}}'",
			},
			TemplateData: map[string]any{
				"Directory": dir,
			},
		}
	}
	MsgErrCreateYumanBinDir = func(dir string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrCreateYumanBinDir",
				Other: "create yuman bin directory '{{.Directory}}'",
			},
			TemplateData: map[string]any{
				"Directory": dir,
			},
		}
	}
	MsgErrYumanMissingBinary = func(path string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrYumanMissingBinary",
				Other: "yuman binary not found in package: '{{.Path}}'",
			},
			TemplateData: map[string]any{
				"Path": path,
			},
		}
	}
	MsgErrInstallYumanPackage = func(dir string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrInstallYumanPackage",
				Other: "install yuman package into '{{.Directory}}'",
			},
			TemplateData: map[string]any{
				"Directory": dir,
			},
		}
	}
	MsgErrChmodYumanBinary = func(path string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrChmodYumanBinary",
				Other: "make '{{.Path}}' executable",
			},
			TemplateData: map[string]any{
				"Path": path,
			},
		}
	}
	MsgErrCreateYumanLink = func(link string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrCreateYumanLink",
				Other: "create yuman symlink '{{.Link}}'",
			},
			TemplateData: map[string]any{
				"Link": link,
			},
		}
	}
	MsgErrRemoveYumanPackage = func(path string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrRemoveYumanPackage",
				Other: "remove yuman package '{{.Path}}'",
			},
			TemplateData: map[string]any{
				"Path": path,
			},
		}
	}
)

// Release describes the latest yuman release selected for this machine.
type Release struct {
	Version   string
	Timestamp string
	Arch      string
	FileName  string
	AssetURL  string
	FileSize  int64
}

// Package is an installed yuman package directory under ~/.local/lib.
type Package struct {
	Name      string
	Path      string
	Version   string
	Timestamp string
	Arch      string
}

// layout holds the yuman install directories. yuman is installed for the
// current user only, so root privileges and the XDG environment are never
// involved: the layout is always under $HOME.
type layout struct {
	libHome string
	binHome string
}

// LatestRelease fetches the latest release from ReleaseRepo and selects the
// linux package that matches this machine's architecture.
func LatestRelease(ctx context.Context) (Release, error) {
	fetched, err := release.Fetch(ctx, ReleaseRepo)
	if err != nil {
		if errors.Is(err, release.ErrNoRelease) {
			return Release{}, tr.LocalizeError(&MsgErrNoYumanRelease)
		}
		return Release{}, tr.WrapError(MsgErrFetchYumanRelease(release.APIURL(ReleaseRepo)), err)
	}

	asset, err := selectAsset(fetched.Assets, release.MachineArch(release.HostArch))
	if err != nil {
		return Release{}, err
	}
	version, timestamp, arch, _ := assetName(asset.Name)
	return Release{
		Version:   version,
		Timestamp: timestamp,
		Arch:      arch,
		FileName:  asset.Name,
		AssetURL:  asset.BrowserDownloadURL,
		FileSize:  asset.Size,
	}, nil
}

// Update installs the latest yuman release into the user directories.
//
// The package the running yuman was loaded from decides what happens: when it
// is a package under $HOME/.local/lib it is compared with the latest release and
// replaced, otherwise this is treated as a first install. Every version lives in
// its own yuman-v<version>-<timestamp>-<arch> directory, and
// $HOME/.local/bin/yuman points at its yuman binary.
func Update(ctx context.Context, verbose bool) (err error) {
	dirs, err := resolveLayout()
	if err != nil {
		return tr.WrapError(&MsgErrResolveYumanLayout, err)
	}
	if verbose {
		fmt.Println(tr.Localize(MsgYumanLayout(dirs.libHome, dirs.binHome)))
	}
	if exe, exeErr := os.Executable(); exeErr == nil && !isManaged(dirs.binHome, exe) {
		fmt.Println(tr.Localize(MsgYumanWarnNotManaged(exe, filepath.Join(dirs.binHome, binaryName))))
	}

	fmt.Println(tr.Localize(&MsgYumanUpgradeChecking))
	latest, err := LatestRelease(ctx)
	if err != nil {
		return
	}
	if verbose {
		fmt.Println(tr.Localize(MsgYumanRelease(
			latest.Version, latest.Timestamp, latest.FileName, release.FormatSize(latest.FileSize),
		)))
	}

	packages, err := listPackages(dirs.libHome)
	if err != nil {
		return tr.WrapError(MsgErrListYumanPackages(dirs.libHome), err)
	}
	if verbose {
		printPackages(packages)
	}

	current, found := runningPackage(dirs.libHome)
	var updateLink bool // whether binary link should be updated specially.
	if len(packages) != 0 {
		latest := &packages[0]
		if !found || current.Path != latest.Path {
			current = *latest
			updateLink = true
		}
	}
	if current.Version != "" {
		if CompareReleases(current, latest) >= 0 {
			fmt.Println(tr.Localize(MsgYumanUpToDate(current.Version, current.Timestamp)))
			// latest version already installed, update link only, if necessary.
			if updateLink {
				return linkBinary(dirs.binHome, filepath.Base(current.Path))
			}
			return nil
		}
		fmt.Println(tr.Localize(MsgYumanUpdateAvailable(
			latest.Version, latest.Timestamp, release.FormatSize(latest.FileSize),
			current.Version, current.Timestamp,
		)))
	}

	return install(ctx, dirs, latest, current, found, packages, verbose)
}

// install downloads latest and installs it, replacing the package the running
// yuman was loaded from.
func install(ctx context.Context, dirs layout, latest Release, current Package, currentFound bool, packages []Package, verbose bool) (err error) {
	pkgPath, err := release.Download(ctx, latest.AssetURL)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(pkgPath)
	}()

	return installPackage(ctx, dirs, latest, pkgPath, current, currentFound, packages, verbose)
}

// installPackage unpacks the downloaded archive into a new lib directory, makes
// its yuman binary executable, points the bin symlink at it and finally removes
// the replaced package.
func installPackage(ctx context.Context, dirs layout, latest Release, pkgPath string, current Package, currentFound bool, packages []Package, verbose bool) error {
	fmt.Println(tr.Localize(MsgYumanInstalling(latest.Version, latest.Timestamp, dirs.libHome)))

	if err := os.MkdirAll(dirs.libHome, 0o755); err != nil {
		return tr.WrapError(MsgErrCreateYumanLibDir(dirs.libHome), err)
	}
	if verbose {
		fmt.Println(tr.Localize(MsgYumanExtracting(pkgPath)))
	}

	// Unpack beside the final location so the rename below stays within one
	// filesystem and never leaves a half-written package behind.
	staging, err := os.MkdirTemp(dirs.libHome, ".yuman-install-*")
	if err != nil {
		return tr.WrapError(&MsgErrCreateYumanTempDir, err)
	}
	defer func() {
		_ = os.RemoveAll(staging)
	}()

	if err = release.ExtractTarGz(ctx, pkgPath, staging); err != nil {
		return err
	}
	if err = stagedBinary(staging); err != nil {
		return err
	}

	// The archive root is the package, so the staging directory itself becomes
	// the versioned package directory.
	name := strings.TrimSuffix(latest.FileName, ".tar.gz")
	target := filepath.Join(dirs.libHome, name)
	if err = replaceDir(staging, target); err != nil {
		return tr.WrapError(MsgErrInstallYumanPackage(target), err)
	}
	binary := filepath.Join(target, binaryName)
	if err = os.Chmod(binary, 0o755); err != nil {
		return tr.WrapError(MsgErrChmodYumanBinary(binary), err)
	}

	link := filepath.Join(dirs.binHome, binaryName)
	if verbose {
		fmt.Println(tr.Localize(MsgYumanLinking(link, binary)))
	}
	if err = linkBinary(dirs.binHome, name); err != nil {
		return err
	}
	return removeReplacedPackages(current, currentFound, packages, name, verbose)
}

// resolveLayout returns the yuman install directories under $HOME.
func resolveLayout() (layout, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return layout{}, err
	}
	return layout{
		libHome: filepath.Join(home, ".local", "lib"),
		binHome: filepath.Join(home, ".local", "bin"),
	}, nil
}

// runningPackage returns the installed package the running yuman was loaded
// from. ok is false when the executable does not live in a package directory
// under libHome, which makes this a first install.
func runningPackage(libHome string) (Package, bool) {
	exe, err := os.Executable()
	if err != nil {
		return Package{}, false
	}
	return packageForExecutable(exe, libHome)
}

// packageForExecutable reports the package that holds exe. It resolves symlinks,
// requires exe to sit directly inside a yuman-* directory under libHome and
// verifies that directory still exists; ok is false otherwise.
func packageForExecutable(exe, libHome string) (Package, bool) {
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return Package{}, false
	}
	dir := filepath.Dir(resolved)
	if filepath.Dir(dir) != filepath.Clean(libHome) {
		return Package{}, false
	}
	name := filepath.Base(dir)
	version, timestamp, arch, ok := parseName(name)
	if !ok {
		return Package{}, false
	}
	if _, err := os.Stat(dir); err != nil {
		return Package{}, false
	}
	return Package{
		Name:      name,
		Path:      dir,
		Version:   version,
		Timestamp: timestamp,
		Arch:      arch,
	}, true
}

// isManaged reports whether exe is the executable reached through binHome/yuman.
func isManaged(binHome, exe string) bool {
	return sameFile(filepath.Join(binHome, binaryName), exe)
}

// sameFile reports whether both paths resolve to the same existing file.
func sameFile(a, b string) bool {
	resolvedA, err := filepath.EvalSymlinks(a)
	if err != nil {
		return false
	}
	resolvedB, err := filepath.EvalSymlinks(b)
	if err != nil {
		return false
	}
	return resolvedA == resolvedB
}

// listPackages returns the yuman package directories in libHome, newest version
// first. A missing libHome means nothing is installed; entries that do not parse
// as yuman packages are ignored.
func listPackages(libHome string) ([]Package, error) {
	entries, err := os.ReadDir(libHome)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	packages := []Package{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		version, timestamp, arch, ok := parseName(entry.Name())
		if !ok {
			continue
		}
		packages = append(packages, Package{
			Name:      entry.Name(),
			Path:      filepath.Join(libHome, entry.Name()),
			Version:   version,
			Timestamp: timestamp,
			Arch:      arch,
		})
	}
	slices.SortFunc(packages, func(a, b Package) int {
		if c := release.CompareVersions(a.Version, b.Version); c != 0 {
			return -c
		}
		return -strings.Compare(a.Timestamp, b.Timestamp)
	})
	return packages, nil
}

// parseName parses a yuman package or release name of the form
// yuman-v0.2.0-202609181503-x86_64, returning its version, timestamp and
// architecture. ok is false when the name does not follow that shape; a version
// may itself contain hyphens, as in yuman-v0.2.0-beta.1-202609181503-x86_64.
func parseName(name string) (version, timestamp, arch string, ok bool) {
	if !strings.HasPrefix(name, libPrefix) {
		return "", "", "", false
	}
	parts := strings.Split(strings.TrimPrefix(name, libPrefix), "-")
	for i, part := range parts {
		if len(part) != timestampLen || !isDigits(part) {
			continue
		}
		if i == 0 || i == len(parts)-1 {
			return "", "", "", false
		}
		return strings.Join(parts[:i], "-"), part, strings.Join(parts[i+1:], "-"), true
	}
	return "", "", "", false
}

// assetName parses a release asset name of the form
// yuman-v0.2.0-202609181503-x86_64.tar.gz.
func assetName(name string) (version, timestamp, arch string, ok bool) {
	if !strings.HasSuffix(name, ".tar.gz") {
		return "", "", "", false
	}
	return parseName(strings.TrimSuffix(name, ".tar.gz"))
}

// isDigits reports whether s is non-empty and contains only ASCII digits.
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// selectAsset picks the yuman asset for arch, preferring the highest version
// and then the newest timestamp. It returns an error when no asset matches.
func selectAsset(assets []release.Asset, arch string) (release.Asset, error) {
	var best release.Asset
	var bestVersion, bestTimestamp string
	for _, asset := range assets {
		version, timestamp, assetArch, ok := assetName(asset.Name)
		if !ok || assetArch != arch {
			continue
		}
		if bestVersion == "" || compareRelease(version, timestamp, bestVersion, bestTimestamp) > 0 {
			best, bestVersion, bestTimestamp = asset, version, timestamp
		}
	}
	if bestVersion == "" {
		return release.Asset{}, tr.LocalizeError(MsgErrYumanNoMatchingAsset(arch))
	}
	return best, nil
}

// compareRelease compares a version and timestamp against another, returning
// -1, 0 or 1. The version wins first; equal versions fall back to the timestamp.
func compareRelease(aVersion, aTimestamp, bVersion, bTimestamp string) int {
	if c := release.CompareVersions(aVersion, bVersion); c != 0 {
		return c
	}
	return strings.Compare(aTimestamp, bTimestamp)
}

// CompareReleases compares an installed yuman package against a release,
// returning -1, 0 or 1.
func CompareReleases(installed Package, rel Release) int {
	return compareRelease(installed.Version, installed.Timestamp, rel.Version, rel.Timestamp)
}

// stagedBinary reports whether an unpacked archive holds the yuman binary at its
// root.
func stagedBinary(staging string) error {
	binary := filepath.Join(staging, binaryName)
	info, err := os.Stat(binary)
	if err != nil || !info.Mode().IsRegular() {
		return tr.LocalizeError(MsgErrYumanMissingBinary(binary))
	}
	return nil
}

// replaceDir moves srcDir to target, clearing a leftover entry at target first so
// installing over an existing directory cannot fail.
func replaceDir(srcDir, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	if _, err := os.Lstat(target); err == nil {
		if err := os.RemoveAll(target); err != nil {
			return err
		}
	}
	return os.Rename(srcDir, target)
}

// linkBinary points binHome/yuman at the yuman of the installed package. The
// link is relative so it keeps working when $HOME moves, and it is swapped in
// atomically so no partial link is ever visible.
func linkBinary(binHome, packageName string) error {
	if err := os.MkdirAll(binHome, 0o755); err != nil {
		return tr.WrapError(MsgErrCreateYumanBinDir(binHome), err)
	}

	link := filepath.Join(binHome, binaryName)
	target := filepath.Join("..", "lib", packageName, binaryName)
	tmp := link + ".tmp"
	_ = os.Remove(tmp)
	if err := os.Symlink(target, tmp); err != nil {
		return tr.WrapError(MsgErrCreateYumanLink(link), err)
	}
	if err := os.Rename(tmp, link); err != nil {
		_ = os.Remove(tmp)
		return tr.WrapError(MsgErrCreateYumanLink(link), err)
	}
	return nil
}

// removeReplacedPackages removes the package the running yuman was loaded from
// once the new one is installed. Without one there is nothing to replace, which
// is the first-install case: the running executable may live anywhere, not only
// under libHome. Several packages means an earlier update left packages behind,
// so they are only reported and left in place.
func removeReplacedPackages(current Package, currentFound bool, installed []Package, installedName string, verbose bool) error {
	if !currentFound {
		return nil
	}
	if leftovers := otherPackages(installed, current, installedName); len(leftovers) > 0 {
		names := make([]string, len(leftovers))
		for i, pkg := range leftovers {
			names[i] = pkg.Name
		}
		fmt.Println(tr.Localize(MsgYumanWarnMultiplePackages(strings.Join(names, ", "))))
		return nil
	}
	if verbose {
		fmt.Println(tr.Localize(MsgYumanRemovingPackage(current.Path)))
	}
	if err := os.RemoveAll(current.Path); err != nil {
		return tr.WrapError(MsgErrRemoveYumanPackage(current.Path), err)
	}
	return nil
}

// otherPackages returns the installed packages besides the replaced one and the
// freshly installed one.
func otherPackages(installed []Package, current Package, installedName string) []Package {
	var others []Package
	for _, pkg := range installed {
		if pkg.Path == current.Path || pkg.Name == installedName {
			continue
		}
		others = append(others, pkg)
	}
	return others
}

// printPackages logs the installed packages when verbose.
func printPackages(packages []Package) {
	if len(packages) == 0 {
		fmt.Println(tr.Localize(&MsgYumanNotInstalled))
		return
	}
	names := make([]string, len(packages))
	for i, pkg := range packages {
		names[i] = pkg.Name
	}
	fmt.Println(tr.Localize(MsgYumanInstalledPackages(strings.Join(names, ", "))))
}
