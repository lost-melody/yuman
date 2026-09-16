// Package yumete installs and updates the yumete tool from its GitHub
// releases. yumete is unrelated to the yume engine: it is installed into the
// current user's directories without root privileges or the XDG environment.
package yumete

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

// ReleaseRepo is the GitHub repository that hosts yumete releases.
const ReleaseRepo = "forfudan/yumete"

const (
	// binaryName is the executable shipped inside a yumete package.
	binaryName = "yumete"
	// libPrefix is the prefix of a yumete package directory.
	libPrefix = "yumete-"
	// platform is the only platform yumete packages are installed from.
	platform = "linux"
)

var (
	MsgYumeteUpdateChecking = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeteUpdateChecking",
			Other: "Checking for the latest yumete release...",
		},
	}
	MsgYumeteNotInstalled = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeteNotInstalled",
			Other: "no yumete package is installed",
		},
	}
	MsgErrResolveYumeteLayout = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrResolveYumeteLayout",
			Other: "resolve yumete install layout",
		},
	}
	MsgErrNoYumeteRelease = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrNoYumeteRelease",
			Other: "no yumete release found",
		},
	}
	MsgErrYumetePackageLayout = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrYumetePackageLayout",
			Other: "unexpected yumete package layout: expected a single 'yumete-*' directory",
		},
	}
	MsgErrCreateYumeteTempDir = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrCreateYumeteTempDir",
			Other: "create yumete temp dir",
		},
	}

	MsgYumeteLayout = func(libHome, binHome string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeteLayout",
				Other: "Install directories: lib '{{.Lib}}', bin '{{.Bin}}'",
			},
			TemplateData: map[string]any{
				"Lib": libHome,
				"Bin": binHome,
			},
		}
	}
	MsgYumeteRelease = func(version, arch, fileName, size string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeteRelease",
				Other: "Latest yumete release: {{.Version}} ({{.Arch}}, {{.FileName}}, {{.Size}})",
			},
			TemplateData: map[string]any{
				"Version":  version,
				"Arch":     arch,
				"FileName": fileName,
				"Size":     size,
			},
		}
	}
	MsgYumeteInstalledPackages = func(packages string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeteInstalledPackages",
				Other: "Installed yumete packages: {{.Packages}}",
			},
			TemplateData: map[string]any{
				"Packages": packages,
			},
		}
	}
	MsgYumeteUpToDate = func(version string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeteUpToDate",
				Other: "yumete is already up to date ({{.Version}})",
			},
			TemplateData: map[string]any{
				"Version": version,
			},
		}
	}
	MsgYumeteUpdateAvailable = func(newVersion, currentVersion string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeteUpdateAvailable",
				Other: "Update available: {{.New}} (installed {{.Current}})",
			},
			TemplateData: map[string]any{
				"New":     newVersion,
				"Current": currentVersion,
			},
		}
	}
	MsgYumeteInstalling = func(version, dir string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeteInstalling",
				Other: "Installing yumete {{.Version}} into '{{.Directory}}'...",
			},
			TemplateData: map[string]any{
				"Version":   version,
				"Directory": dir,
			},
		}
	}
	MsgYumeteExtracting = func(pkg string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeteExtracting",
				Other: "Extracting '{{.Package}}'...",
			},
			TemplateData: map[string]any{
				"Package": pkg,
			},
		}
	}
	MsgYumeteLinking = func(link, target string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeteLinking",
				Other: "'{{.Link}}' -> '{{.Target}}'",
			},
			TemplateData: map[string]any{
				"Link":   link,
				"Target": target,
			},
		}
	}
	MsgYumeteRemovingPackage = func(path string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeteRemovingPackage",
				Other: "Removing replaced yumete package '{{.Path}}'",
			},
			TemplateData: map[string]any{
				"Path": path,
			},
		}
	}
	MsgYumeteWarnMultiplePackages = func(packages string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeteWarnMultiplePackages",
				Other: "multiple yumete packages found ({{.Packages}}); skipping removal of old packages",
			},
			TemplateData: map[string]any{
				"Packages": packages,
			},
		}
	}

	MsgErrFetchYumeteRelease = func(url string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrFetchYumeteRelease",
				Other: "fetch latest yumete release from '{{.URL}}'",
			},
			TemplateData: map[string]any{
				"URL": url,
			},
		}
	}
	MsgErrYumeteNoMatchingAsset = func(arch string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrYumeteNoMatchingAsset",
				Other: "no matching linux yumete package for '{{.Arch}}'",
			},
			TemplateData: map[string]any{
				"Arch": arch,
			},
		}
	}
	MsgErrListYumetePackages = func(dir string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrListYumetePackages",
				Other: "list installed yumete packages in '{{.Directory}}'",
			},
			TemplateData: map[string]any{
				"Directory": dir,
			},
		}
	}
	MsgErrCreateYumeteLibDir = func(dir string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrCreateYumeteLibDir",
				Other: "create yumete lib directory '{{.Directory}}'",
			},
			TemplateData: map[string]any{
				"Directory": dir,
			},
		}
	}
	MsgErrCreateYumeteBinDir = func(dir string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrCreateYumeteBinDir",
				Other: "create yumete bin directory '{{.Directory}}'",
			},
			TemplateData: map[string]any{
				"Directory": dir,
			},
		}
	}
	MsgErrYumeteMissingBinary = func(path string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrYumeteMissingBinary",
				Other: "yumete binary not found in package: '{{.Path}}'",
			},
			TemplateData: map[string]any{
				"Path": path,
			},
		}
	}
	MsgErrInstallYumetePackage = func(dir string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrInstallYumetePackage",
				Other: "install yumete package into '{{.Directory}}'",
			},
			TemplateData: map[string]any{
				"Directory": dir,
			},
		}
	}
	MsgErrChmodYumeteBinary = func(path string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrChmodYumeteBinary",
				Other: "make '{{.Path}}' executable",
			},
			TemplateData: map[string]any{
				"Path": path,
			},
		}
	}
	MsgErrCreateYumeteLink = func(link string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrCreateYumeteLink",
				Other: "create yumete symlink '{{.Link}}'",
			},
			TemplateData: map[string]any{
				"Link": link,
			},
		}
	}
	MsgErrRemoveYumetePackage = func(path string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrRemoveYumetePackage",
				Other: "remove yumete package '{{.Path}}'",
			},
			TemplateData: map[string]any{
				"Path": path,
			},
		}
	}
)

// Release describes the latest yumete release selected for this machine.
type Release struct {
	Version  string
	Arch     string
	FileName string
	AssetURL string
	FileSize int64
}

// Package is an installed yumete package directory under ~/.local/lib.
type Package struct {
	Name    string
	Path    string
	Version string
	Arch    string
}

// layout holds the yumete install directories. yumete is installed for the
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
			return Release{}, tr.LocalizeError(&MsgErrNoYumeteRelease)
		}
		return Release{}, tr.WrapError(MsgErrFetchYumeteRelease(release.APIURL(ReleaseRepo)), err)
	}

	asset, err := selectAsset(fetched.Assets, release.MachineArch(release.HostArch))
	if err != nil {
		return Release{}, err
	}
	version, arch, _ := assetName(asset.Name)
	return Release{
		Version:  version,
		Arch:     arch,
		FileName: asset.Name,
		AssetURL: asset.BrowserDownloadURL,
		FileSize: asset.Size,
	}, nil
}

// Update installs the latest yumete release into the user directories.
//
// The packages under $HOME/.local/lib decide what happens: with none installed
// the latest release is installed, otherwise the newest installed version is
// compared against the latest release and replaced when the release is newer.
// Every version lives in its own yumete-<version>-linux-<arch> directory, and
// $HOME/.local/bin/yumete points at the bin/yumete of the current one.
func Update(ctx context.Context, verbose bool) (err error) {
	dirs, err := resolveLayout()
	if err != nil {
		return tr.WrapError(&MsgErrResolveYumeteLayout, err)
	}
	if verbose {
		fmt.Println(tr.Localize(MsgYumeteLayout(dirs.libHome, dirs.binHome)))
	}

	fmt.Println(tr.Localize(&MsgYumeteUpdateChecking))
	latest, err := LatestRelease(ctx)
	if err != nil {
		return
	}
	if verbose {
		fmt.Println(tr.Localize(MsgYumeteRelease(
			latest.Version, latest.Arch, latest.FileName, release.FormatSize(latest.FileSize),
		)))
	}

	packages, err := listPackages(dirs.libHome)
	if err != nil {
		return tr.WrapError(MsgErrListYumetePackages(dirs.libHome), err)
	}
	if verbose {
		printPackages(packages)
	}

	// listPackages sorts newest first, so the head is the version the update
	// would replace.
	if len(packages) > 0 {
		current := packages[0]
		if release.CompareVersions(current.Version, latest.Version) >= 0 {
			fmt.Println(tr.Localize(MsgYumeteUpToDate(current.Version)))
			return nil
		}
		fmt.Println(tr.Localize(MsgYumeteUpdateAvailable(latest.Version, current.Version)))
	}

	return install(ctx, dirs, latest, packages, verbose)
}

// install downloads latest and installs it, replacing the packages that were
// installed before.
func install(ctx context.Context, dirs layout, latest Release, packages []Package, verbose bool) (err error) {
	pkgPath, err := release.Download(ctx, latest.AssetURL)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(pkgPath)
	}()

	return installPackage(ctx, dirs, latest, pkgPath, packages, verbose)
}

// installPackage unpacks the downloaded package at pkgPath into the lib
// directory, makes bin/yumete executable, points the bin symlink at the new
// package and finally removes the replaced package.
func installPackage(ctx context.Context, dirs layout, latest Release, pkgPath string, packages []Package, verbose bool) error {
	fmt.Println(tr.Localize(MsgYumeteInstalling(latest.Version, dirs.libHome)))

	if err := os.MkdirAll(dirs.libHome, 0o755); err != nil {
		return tr.WrapError(MsgErrCreateYumeteLibDir(dirs.libHome), err)
	}
	if verbose {
		fmt.Println(tr.Localize(MsgYumeteExtracting(pkgPath)))
	}

	// Unpack beside the final location so the rename below stays within one
	// filesystem and never leaves a half-written package behind.
	staging, err := os.MkdirTemp(dirs.libHome, ".yumete-install-*")
	if err != nil {
		return tr.WrapError(&MsgErrCreateYumeteTempDir, err)
	}
	defer func() {
		_ = os.RemoveAll(staging)
	}()

	if err = release.ExtractTarGz(ctx, pkgPath, staging); err != nil {
		return err
	}
	name, srcDir, err := stagedPackage(staging)
	if err != nil {
		return err
	}

	target := filepath.Join(dirs.libHome, name)
	if err = replaceDir(srcDir, target); err != nil {
		return tr.WrapError(MsgErrInstallYumetePackage(target), err)
	}
	binary := filepath.Join(target, "bin", binaryName)
	if err = os.Chmod(binary, 0o755); err != nil {
		return tr.WrapError(MsgErrChmodYumeteBinary(binary), err)
	}

	link := filepath.Join(dirs.binHome, binaryName)
	if verbose {
		fmt.Println(tr.Localize(MsgYumeteLinking(link, binary)))
	}
	if err = linkBinary(dirs.binHome, name); err != nil {
		return err
	}
	return removeReplacedPackages(packages, verbose)
}

// resolveLayout returns the yumete install directories under $HOME.
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

// listPackages returns the yumete package directories in libHome, newest
// version first. A missing libHome means nothing is installed; entries that do
// not parse as yumete packages are ignored.
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
		version, arch, ok := parseName(entry.Name())
		if !ok {
			continue
		}
		packages = append(packages, Package{
			Name:    entry.Name(),
			Path:    filepath.Join(libHome, entry.Name()),
			Version: version,
			Arch:    arch,
		})
	}
	slices.SortFunc(packages, func(a, b Package) int {
		return -release.CompareVersions(a.Version, b.Version)
	})
	return packages, nil
}

// parseName parses a yumete package name of the form
// yumete-0.1.0-linux-x86_64, returning its version and architecture. ok is
// false when the name does not follow that shape; a version may itself contain
// hyphens, as in yumete-0.1.0-beta.1-linux-x86_64.
func parseName(name string) (version, arch string, ok bool) {
	if !strings.HasPrefix(name, libPrefix) {
		return "", "", false
	}
	parts := strings.Split(strings.TrimPrefix(name, libPrefix), "-")
	for i, part := range parts {
		if part != platform {
			continue
		}
		if i == 0 || i == len(parts)-1 {
			return "", "", false
		}
		return strings.Join(parts[:i], "-"), strings.Join(parts[i+1:], "-"), true
	}
	return "", "", false
}

// assetName parses a release asset name of the form
// yumete-0.1.0-linux-x86_64.tar.gz, returning its version and architecture.
func assetName(name string) (version, arch string, ok bool) {
	if !strings.HasSuffix(name, ".tar.gz") {
		return "", "", false
	}
	return parseName(strings.TrimSuffix(name, ".tar.gz"))
}

// selectAsset picks the yumete asset for arch, preferring the highest version.
// It returns an error when no asset matches.
func selectAsset(assets []release.Asset, arch string) (release.Asset, error) {
	var best release.Asset
	var bestVersion string
	for _, asset := range assets {
		version, assetArch, ok := assetName(asset.Name)
		if !ok || assetArch != arch {
			continue
		}
		if bestVersion == "" || release.CompareVersions(version, bestVersion) > 0 {
			best, bestVersion = asset, version
		}
	}
	if bestVersion == "" {
		return release.Asset{}, tr.LocalizeError(MsgErrYumeteNoMatchingAsset(arch))
	}
	return best, nil
}

// stagedPackage locates the single yumete-* directory inside an unpacked
// release, returning its name and path. It also requires the package to hold a
// yumete executable under bin/.
func stagedPackage(staging string) (name, dir string, err error) {
	entries, err := os.ReadDir(staging)
	if err != nil {
		return "", "", err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	if len(names) != 1 {
		return "", "", tr.LocalizeError(&MsgErrYumetePackageLayout)
	}
	name = names[0]
	if _, _, ok := parseName(name); !ok {
		return "", "", tr.LocalizeError(&MsgErrYumetePackageLayout)
	}

	dir = filepath.Join(staging, name)
	binary := filepath.Join(dir, "bin", binaryName)
	if info, statErr := os.Stat(binary); statErr != nil || !info.Mode().IsRegular() {
		return "", "", tr.LocalizeError(MsgErrYumeteMissingBinary(binary))
	}
	return name, dir, nil
}

// replaceDir moves srcDir to target, clearing a leftover entry at target first
// so installing over an existing directory cannot fail.
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

// linkBinary points binHome/yumete at the bin/yumete of the installed package.
// The link is relative so it keeps working when $HOME moves, and it is swapped
// in atomically so no partial link is ever visible.
func linkBinary(binHome, packageName string) error {
	if err := os.MkdirAll(binHome, 0o755); err != nil {
		return tr.WrapError(MsgErrCreateYumeteBinDir(binHome), err)
	}

	link := filepath.Join(binHome, binaryName)
	target := filepath.Join("..", "lib", packageName, "bin", binaryName)
	tmp := link + ".tmp"
	_ = os.Remove(tmp)
	if err := os.Symlink(target, tmp); err != nil {
		return tr.WrapError(MsgErrCreateYumeteLink(link), err)
	}
	if err := os.Rename(tmp, link); err != nil {
		_ = os.Remove(tmp)
		return tr.WrapError(MsgErrCreateYumeteLink(link), err)
	}
	return nil
}

// removeReplacedPackages removes the package that the install replaced. A
// single old package is unambiguous and removed; several means an earlier
// update left packages behind, so they are only reported and left in place.
func removeReplacedPackages(replaced []Package, verbose bool) error {
	switch len(replaced) {
	case 0:
		return nil
	case 1:
		if verbose {
			fmt.Println(tr.Localize(MsgYumeteRemovingPackage(replaced[0].Path)))
		}
		if err := os.RemoveAll(replaced[0].Path); err != nil {
			return tr.WrapError(MsgErrRemoveYumetePackage(replaced[0].Path), err)
		}
		return nil
	default:
		names := make([]string, len(replaced))
		for i, pkg := range replaced {
			names[i] = pkg.Name
		}
		fmt.Println(tr.Localize(MsgYumeteWarnMultiplePackages(strings.Join(names, ", "))))
		return nil
	}
}

// printPackages logs the installed packages when verbose.
func printPackages(packages []Package) {
	if len(packages) == 0 {
		fmt.Println(tr.Localize(&MsgYumeteNotInstalled))
		return
	}
	names := make([]string, len(packages))
	for i, pkg := range packages {
		names[i] = pkg.Name
	}
	fmt.Println(tr.Localize(MsgYumeteInstalledPackages(strings.Join(names, ", "))))
}
