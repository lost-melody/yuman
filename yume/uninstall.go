package yume

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// libYumeName is the fcitx5 addon library shipped by a yume release.
const libYumeName = "libyume.so"

// systemShareDir is where a system install puts the fcitx5 configuration.
const systemShareDir = "/usr/share"

// systemLocaleDir is the system gettext directory; fcitx5 does not load
// translations from user directories. It is a variable so tests can point it
// at a temporary directory.
var systemLocaleDir = "/usr/share/locale"

// uninstallPlan lists the paths one uninstall pass removes: files are removed
// with rm -f, whole directories with rm -rf, and tidyDirs with rmdir once they
// are left empty.
type uninstallPlan struct {
	files    []string
	dirs     []string
	tidyDirs []string
}

var (
	MsgErrUninstallFailed = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrUninstallFailed",
			Other: "uninstall failed",
		},
	}
	MsgErrResolveUninstallLayout = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrResolveUninstallLayout",
			Other: "resolve uninstall layout",
		},
	}
	MsgErrSystemUninstallPkexec = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrSystemUninstallPkexec",
			Other: "system uninstall requires pkexec",
		},
	}
	MsgWarnUninstallSkipped = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "WarnUninstallSkipped",
			Other: "some files were not removed (pkexec was declined)",
		},
	}

	MsgUninstallingYume = func(target string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "UninstallingYume",
				Other: "Uninstalling yume from {{.Target}}...",
			},
			TemplateData: map[string]any{
				"Target": target,
			},
		}
	}
	MsgRemovingFile = func(path string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "RemovingFile",
				Other: "Removing '{{.Path}}'",
			},
			TemplateData: map[string]any{
				"Path": path,
			},
		}
	}
)

// UninstallYume removes yume from user directories or system directories,
// undoing what InstallYume wrote for the matching mode.
//
// The yume data directory (~/.local/share/yume) holds user data, so it is
// never removed automatically; pass purge to remove it as well.
//
// User uninstalls remove the addon library, the fcitx5 configuration and the
// environment.d entry directly. The system locale files (fcitx5 never loads
// them from user directories) are removed through pkexec as a best-effort
// step, together with a purged data directory that is root-owned because a
// system install created it. System uninstalls remove the system files and
// the locale files, and the data directory when purging, in a single pkexec
// command.
func UninstallYume(ctx context.Context, userDirs, purge bool, verbose bool) (err error) {
	target := tr.Localize(&MsgInstallUserDirs)
	if !userDirs {
		target = tr.Localize(&MsgInstallSystemDirs)
	}
	fmt.Println(tr.Localize(MsgUninstallingYume(target)))

	home, err := os.UserHomeDir()
	if err != nil {
		return wrapError(&MsgErrResolveUninstallLayout, err)
	}
	if userDirs {
		return uninstallUserDirs(ctx, home, purge, verbose)
	}
	return uninstallSystemDirs(ctx, home, purge, verbose)
}

// uninstallUserDirs removes a user install: the user files first, then the
// purged data directory, and finally the system locale files.
func uninstallUserDirs(ctx context.Context, home string, purge bool, verbose bool) error {
	if err := runUninstallPlan(ctx, planUserUninstall(home), false, verbose); err != nil {
		return err
	}

	var bestEffort *uninstallPlan
	if purge {
		// The data directory may be root-owned because a system install
		// created it through pkexec; retry it through pkexec when the user
		// cannot remove it directly.
		dataPlan := uninstallPlan{dirs: []string{yumeDataDir(home)}}
		if err := runUninstallPlan(ctx, dataPlan, false, verbose); err != nil {
			bestEffort = &dataPlan
		}
	}

	if mos := localeMoFiles(systemLocaleDir); len(mos) > 0 {
		localePlan := uninstallPlan{files: mos, tidyDirs: localeTidyDirs(mos)}
		if bestEffort == nil {
			bestEffort = &localePlan
		} else {
			bestEffort.files = append(bestEffort.files, localePlan.files...)
			bestEffort.tidyDirs = append(bestEffort.tidyDirs, localePlan.tidyDirs...)
		}
	}
	if bestEffort != nil {
		if err := runUninstallPlan(ctx, *bestEffort, true, verbose); err != nil {
			// Best-effort: a declined pkexec prompt must not abort the uninstall.
			fmt.Printf("%s\n", tr.Localize(&MsgWarnUninstallSkipped))
		}
	}
	return nil
}

// uninstallSystemDirs removes a system install with a single pkexec command.
func uninstallSystemDirs(ctx context.Context, home string, purge bool, verbose bool) error {
	plan := planSystemUninstall(home, globLibYumeFiles(), localeMoFiles(systemLocaleDir), purge)
	return runUninstallPlan(ctx, plan, true, verbose)
}

// yumeDataDir returns the yume data directory; yume data always lives there
// regardless of the install mode.
func yumeDataDir(home string) string {
	return filepath.Join(home, ".local", "share", "yume")
}

// ExistingYumeDataDir returns the yume user data directory when it exists.
// The directory holds user data such as dictionaries, so it must only be
// removed after an explicit purge request.
func ExistingYumeDataDir() (string, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}
	dir := yumeDataDir(home)
	if _, err := os.Stat(dir); err != nil {
		return "", false
	}
	return dir, true
}

// planUserUninstall lists the files a user install wrote. The yume data
// directory is planned separately because removing it may need pkexec.
func planUserUninstall(home string) uninstallPlan {
	return uninstallPlan{
		files: []string{
			filepath.Join(home, ".local", "lib", "fcitx5", libYumeName),
			filepath.Join(home, ".local", "share", "fcitx5", "addon", "yume.conf"),
			filepath.Join(home, ".local", "share", "fcitx5", "inputmethod", "yume.conf"),
			filepath.Join(home, ".config", "environment.d", environmentFileName),
		},
		tidyDirs: []string{
			filepath.Join(home, ".local", "lib", "fcitx5"),
			filepath.Join(home, ".local", "share", "fcitx5", "addon"),
			filepath.Join(home, ".local", "share", "fcitx5", "inputmethod"),
			filepath.Join(home, ".local", "share", "fcitx5"),
			filepath.Join(home, ".config", "environment.d"),
		},
	}
}

// planSystemUninstall lists the files a system install wrote. libYumes are the
// existing addon library locations and mos the existing gettext catalogs,
// both normally discovered by globbing the system directories. The yume data
// directory holds user data and is only included when purging.
func planSystemUninstall(home string, libYumes, mos []string, purge bool) uninstallPlan {
	plan := uninstallPlan{
		files: append(append([]string{}, libYumes...),
			filepath.Join(systemShareDir, "fcitx5", "addon", "yume.conf"),
			filepath.Join(systemShareDir, "fcitx5", "inputmethod", "yume.conf"),
		),
	}
	if purge {
		// A system install writes the yume data directory as root.
		plan.dirs = []string{yumeDataDir(home)}
	}
	plan.files = append(plan.files, mos...)
	for _, lib := range libYumes {
		plan.tidyDirs = append(plan.tidyDirs, filepath.Dir(lib))
	}
	plan.tidyDirs = append(plan.tidyDirs,
		filepath.Join(systemShareDir, "fcitx5", "addon"),
		filepath.Join(systemShareDir, "fcitx5", "inputmethod"),
		filepath.Join(systemShareDir, "fcitx5"),
	)
	plan.tidyDirs = append(plan.tidyDirs, localeTidyDirs(mos)...)
	return plan
}

// globLibYumeFiles returns every yume addon library installed under the
// system fcitx5 library directories, mirroring the locations
// findSystemLibParent picks from.
func globLibYumeFiles() []string {
	var found []string
	for _, pattern := range []string{
		"/usr/lib/*/fcitx5/" + libYumeName,
		"/usr/lib/fcitx5/" + libYumeName,
	} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		found = append(found, matches...)
	}
	return found
}

// localeMoFiles returns the yume gettext catalogs installed under localeDir.
func localeMoFiles(localeDir string) []string {
	pattern := filepath.Join(localeDir, "*", "LC_MESSAGES", yumeLocaleDomain+".mo")
	matches, _ := filepath.Glob(pattern)
	return matches
}

// localeTidyDirs returns the locale directories that may be left empty once
// the given catalogs are removed.
func localeTidyDirs(mos []string) []string {
	dirs := make([]string, 0, len(mos)*2)
	for _, mo := range mos {
		messages := filepath.Dir(mo)
		dirs = append(dirs, messages, filepath.Dir(messages))
	}
	return dirs
}

// runUninstallPlan removes the plan's paths through a single shell command.
// The plan is skipped when none of its removal targets exist, so an uninstall
// never asks for root privileges just to remove nothing. tidyDirs that do not
// exist are dropped because rmdir would fail on them.
func runUninstallPlan(ctx context.Context, plan uninstallPlan, usePkexec bool, verbose bool) error {
	targets := existingPaths(append(append([]string{}, plan.files...), plan.dirs...))
	if len(targets) == 0 {
		return nil
	}
	tidyDirs := existingPaths(plan.tidyDirs)
	if verbose {
		for _, path := range targets {
			fmt.Printf("  %s\n", tr.Localize(MsgRemovingFile(path)))
		}
	}
	script := buildRemovalScript(plan.files, plan.dirs, tidyDirs)
	return runShellScript(ctx, script, usePkexec, &MsgErrSystemUninstallPkexec, &MsgErrUninstallFailed)
}

// buildRemovalScript assembles removal operations into a single "&&"-joined
// shell command: files are removed with rm -f (missing files are tolerated),
// directories with rm -rf, and tidyDirs with rmdir only when they are empty.
func buildRemovalScript(files, dirs, tidyDirs []string) string {
	var parts []string
	if len(files) > 0 {
		parts = append(parts, "rm -f "+strings.Join(shellQuoteAll(files), " "))
	}
	if len(dirs) > 0 {
		parts = append(parts, "rm -rf "+strings.Join(shellQuoteAll(dirs), " "))
	}
	if len(tidyDirs) > 0 {
		parts = append(parts, "rmdir --ignore-fail-on-non-empty "+strings.Join(shellQuoteAll(tidyDirs), " "))
	}
	return strings.Join(parts, " && ")
}

// existingPaths returns the paths that exist on disk, preserving order.
func existingPaths(paths []string) []string {
	var existing []string
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existing = append(existing, path)
		}
	}
	return existing
}
