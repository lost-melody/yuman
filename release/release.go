// Package release fetches GitHub release metadata, downloads release assets
// and unpacks their tarballs. It holds the parts of installing a release that
// are independent of the program being installed.
package release

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrNoRelease reports a repository that has no published release.
var ErrNoRelease = errors.New("no release found")

// Asset is the subset of a GitHub release asset that is needed.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Release is the subset of the GitHub releases/latest response that is needed.
type Release struct {
	Assets []Asset `json:"assets"`
}

// APIURL returns the GitHub API endpoint for a repository's latest release.
func APIURL(repo string) string {
	return "https://api.github.com/repos/" + repo + "/releases/latest"
}

// Fetch queries the GitHub API for the latest release of repo. A repository
// without a published release yields ErrNoRelease.
func Fetch(ctx context.Context, repo string) (Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, APIURL(repo), nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return Release{}, ErrNoRelease
	}
	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("unexpected status %s", resp.Status)
	}

	var release Release
	if err = json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Release{}, err
	}
	return release, nil
}
