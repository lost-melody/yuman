package yume

import (
	"context"
	"errors"
	"strings"

	"github.com/lost-melody/yuman/release"
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// ReleaseRepo is the GitHub repository that hosts yume releases.
const ReleaseRepo = "forfudan/yume-release"

var (
	MsgErrFetchRelease = func(url string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrFetchRelease",
				Other: "fetch latest yume release from '{{.URL}}'",
			},
			TemplateData: map[string]any{"URL": url},
		}
	}
	MsgErrNoRelease = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrNoRelease",
			Other: "no yume release found",
		},
	}
	MsgErrNoMatchingAsset = func(arch string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrNoMatchingAsset",
				Other: "no matching linux package for {{.Arch}}",
			},
			TemplateData: map[string]any{"Arch": arch},
		}
	}
)

// YumeRelease describes the latest yume release selected for this machine.
type YumeRelease struct {
	Version   string
	Timestamp string
	Arch      string
	FileName  string
	AssetURL  string
	FileSize  int64
}

// LatestYumeRelease fetches the latest release from ReleaseRepo and selects
// the newest linux package that matches this machine's architecture.
func LatestYumeRelease(ctx context.Context) (YumeRelease, error) {
	fetched, err := release.Fetch(ctx, ReleaseRepo)
	if err != nil {
		if errors.Is(err, release.ErrNoRelease) {
			return YumeRelease{}, tr.LocalizeError(&MsgErrNoRelease)
		}
		return YumeRelease{}, tr.WrapError(MsgErrFetchRelease(release.APIURL(ReleaseRepo)), err)
	}

	asset, err := selectReleaseAsset(fetched.Assets, release.MachineArch(release.HostArch))
	if err != nil {
		return YumeRelease{}, err
	}
	version, timestamp, arch, _ := parseReleaseAssetName(asset.Name)
	return YumeRelease{
		Version:   version,
		Timestamp: timestamp,
		Arch:      arch,
		FileName:  asset.Name,
		AssetURL:  asset.BrowserDownloadURL,
		FileSize:  asset.Size,
	}, nil
}

// selectReleaseAsset picks the newest linux asset for arch. Every asset in a
// release shares the same version, so only arch and timestamp are compared;
// the greatest timestamp wins. It returns an error when no asset matches.
func selectReleaseAsset(assets []release.Asset, arch string) (release.Asset, error) {
	var best release.Asset
	var bestTimestamp string
	for _, asset := range assets {
		_, timestamp, a, ok := parseReleaseAssetName(asset.Name)
		if !ok || a != arch {
			continue
		}
		if bestTimestamp == "" || timestamp > bestTimestamp {
			best = asset
			bestTimestamp = timestamp
		}
	}
	if bestTimestamp == "" {
		return release.Asset{}, tr.LocalizeError(MsgErrNoMatchingAsset(arch))
	}
	return best, nil
}

// parseReleaseAssetName parses a linux release asset name of the form
// Yume-v3.12.0-20260826130405-linux-x86_64.tar.gz, returning the version,
// timestamp and arch. ok is false when the name does not follow that shape.
//
// The anchor is the "linux" segment, not the timestamp, because a version may
// itself contain hyphens and may join the timestamp with a dot: a pre-release
// is published as Yume-v3.13.0-beta.20261018122440-linux-x86_64.tar.gz, where
// splitting on "-" never yields a bare 14-digit field. Anchoring on the
// timestamp made every such release invisible to `yuman yume update`, which
// reported "no matching linux package" with the asset sitting right there.
// This is also how the sibling parser in the yumete package reads its names.
func parseReleaseAssetName(name string) (version, timestamp, arch string, ok bool) {
	if !strings.HasSuffix(name, ".tar.gz") {
		return "", "", "", false
	}
	const prefix = "Yume-"
	if !strings.HasPrefix(name, prefix) {
		return "", "", "", false
	}
	rest := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".tar.gz")

	parts := strings.Split(rest, "-")
	for i, part := range parts {
		if part != "linux" {
			continue
		}
		// The layout is "<version-carrying-the-timestamp>-linux-<arch>".
		if i == 0 || i == len(parts)-1 {
			return "", "", "", false
		}
		version, timestamp, ok = splitTimestamp(strings.Join(parts[:i], "-"))
		if !ok {
			return "", "", "", false
		}
		return version, timestamp, strings.Join(parts[i+1:], "-"), true
	}
	return "", "", "", false
}

// splitTimestamp takes the 14-digit build stamp off the end of the version
// text an asset name carries, returning what precedes it and the stamp.
//
// Both separators are in use: a release names it "v3.12.0-20260826130405" and
// a pre-release "v3.13.0-beta.20261018122440". ok is false when the text does
// not end in a stamp, which is what makes a name without one no asset at all.
func splitTimestamp(text string) (version, timestamp string, ok bool) {
	if len(text) < timestampLen+1 {
		return "", "", false
	}
	timestamp = text[len(text)-timestampLen:]
	if !isDigits(timestamp) {
		return "", "", false
	}
	version = text[:len(text)-timestampLen]
	switch version[len(version)-1] {
	case '-', '.':
		return version[:len(version)-1], timestamp, true
	default:
		return "", "", false
	}
}

const timestampLen = 14

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

// CompareReleases compares an installed yume against a release, returning -1,
// 0 or 1. Version wins first; when versions are equal the timestamps decide,
// with the installed build timestamp compared to the release timestamp.
func CompareReleases(installed YumeVersion, rel YumeRelease) int {
	if c := release.CompareVersions(installed.Version, rel.Version); c != 0 {
		return c
	}
	return strings.Compare(installed.Build, rel.Timestamp)
}
