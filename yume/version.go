package yume

import (
	"os"
	"path/filepath"
	"strings"
)

// systemYumeDir is where a system install puts the yume data directory,
// including its VERSION file. It is a variable so tests can redirect it.
var systemYumeDir = "/usr/share/yume"

// YumeVersion holds the version, build and arch metadata parsed from a
// share/yume/VERSION file.
type YumeVersion struct {
	Version string `json:"version"`
	Build   string `json:"build"`
	Arch    string `json:"arch"`
}

// parseVersion parses the version=, build= and arch= fields from VERSION
// content, ignoring any other lines.
func parseVersion(content string) YumeVersion {
	var v YumeVersion
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "version="):
			v.Version = strings.TrimSpace(strings.TrimPrefix(line, "version="))
		case strings.HasPrefix(line, "build="):
			v.Build = strings.TrimSpace(strings.TrimPrefix(line, "build="))
		case strings.HasPrefix(line, "arch="):
			v.Arch = strings.TrimSpace(strings.TrimPrefix(line, "arch="))
		}
	}
	return v
}

// versionFromFile reads and parses the VERSION file at path.
func versionFromFile(path string) (YumeVersion, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return YumeVersion{}, err
	}
	return parseVersion(string(data)), nil
}

// PackageVersion reads the package's share/yume/VERSION file under srcDir.
func PackageVersion(srcDir string) (YumeVersion, error) {
	return versionFromFile(filepath.Join(srcDir, "share", "yume", "VERSION"))
}

// InstalledYumeVersion returns the installed yume version, preferring the user
// installation and falling back to the system installation. found reports
// whether a VERSION file was located at all.
func InstalledYumeVersion() (v YumeVersion, found bool) {
	dataHome, err := userDataHome()
	if err != nil {
		return YumeVersion{}, false
	}
	for _, path := range []string{
		filepath.Join(dataHome, "yume", "VERSION"),
		filepath.Join(systemYumeDir, "VERSION"),
	} {
		v, err = versionFromFile(path)
		if err == nil {
			return v, true
		}
	}
	return YumeVersion{}, false
}

// Lines renders the version metadata as version=, build= and arch= lines,
// mirroring the VERSION file format.
func (v YumeVersion) Lines() []string {
	return []string{
		"version=" + v.Version,
		"build=" + v.Build,
		"arch=" + v.Arch,
	}
}
