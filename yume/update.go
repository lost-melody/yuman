package yume

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

// releaseAsset is the subset of a GitHub release asset we need.
type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// githubRelease is the subset of the GitHub releases/latest response we need.
type githubRelease struct {
	Assets []releaseAsset `json:"assets"`
}

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
	url := releaseAPIURL()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return YumeRelease{}, wrapError(MsgErrFetchRelease(url), err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return YumeRelease{}, wrapError(MsgErrFetchRelease(url), err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return YumeRelease{}, tr.LocalizeError(&MsgErrNoRelease)
	}
	if resp.StatusCode != http.StatusOK {
		return YumeRelease{}, wrapError(MsgErrFetchRelease(url), fmt.Errorf("unexpected status %s", resp.Status))
	}

	var release githubRelease
	if err = json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return YumeRelease{}, wrapError(MsgErrFetchRelease(url), err)
	}

	asset, err := selectReleaseAsset(release.Assets, machineArch(hostArch))
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

// releaseAPIURL returns the GitHub API endpoint for the latest release.
func releaseAPIURL() string {
	return "https://api.github.com/repos/" + ReleaseRepo + "/releases/latest"
}

// selectReleaseAsset picks the newest linux asset for arch. Every asset in a
// release shares the same version, so only arch and timestamp are compared;
// the greatest timestamp wins. It returns an error when no asset matches.
func selectReleaseAsset(assets []releaseAsset, arch string) (releaseAsset, error) {
	var best releaseAsset
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
		return releaseAsset{}, tr.LocalizeError(MsgErrNoMatchingAsset(arch))
	}
	return best, nil
}

// parseReleaseAssetName parses a linux release asset name of the form
// Yume-v3.12.0-20260826130405-linux-x86_64.tar.gz, returning the version,
// timestamp and arch. ok is false when the name does not follow that shape.
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
		if len(part) != timestampLen || !isDigits(part) {
			continue
		}
		// The layout after the timestamp is "-linux-<arch>".
		tail := parts[i+1:]
		if len(tail) < 2 || tail[0] != "linux" {
			return "", "", "", false
		}
		return strings.Join(parts[:i], "-"), part, strings.Join(tail[1:], "-"), true
	}
	return "", "", "", false
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

// CompareVersions compares two version strings, ignoring an optional leading
// "v" and comparing dot-separated numeric segments. It returns -1, 0 or 1.
func CompareVersions(a, b string) int {
	a = strings.TrimPrefix(strings.TrimPrefix(a, "v"), "V")
	b = strings.TrimPrefix(strings.TrimPrefix(b, "v"), "V")
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		var an, bn int
		if i < len(as) {
			an = leadingInt(as[i])
		}
		if i < len(bs) {
			bn = leadingInt(bs[i])
		}
		switch {
		case an < bn:
			return -1
		case an > bn:
			return 1
		}
	}
	return 0
}

// CompareReleases compares an installed yume against a release, returning -1,
// 0 or 1. Version wins first; when versions are equal the timestamps decide,
// with the installed build timestamp compared to the release timestamp.
func CompareReleases(installed YumeVersion, release YumeRelease) int {
	if c := CompareVersions(installed.Version, release.Version); c != 0 {
		return c
	}
	return strings.Compare(installed.Build, release.Timestamp)
}

// FormatSize returns the fileSize in the "1.23 MiB" format.
func FormatSize(fileSize int64) string {
	const unit = 1024
	if fileSize <= 0 {
		return "0 B"
	}
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	size := float64(fileSize)
	exp := 0
	for size >= unit && exp < len(units)-1 {
		size /= unit
		exp++
	}
	switch exp {
	case 0:
		return fmt.Sprintf("%.0f %s", size, units[exp])
	default:
		return fmt.Sprintf("%.2f %s", size, units[exp])
	}
}

// leadingInt returns the leading run of ASCII digits in s as an integer.
func leadingInt(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			break
		}
		n = n*10 + int(s[i]-'0')
	}
	return n
}
