package release

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var (
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
)

// Download fetches the asset at url into a temporary .tar.gz file and returns
// its path. The caller owns the file and must remove it.
func Download(ctx context.Context, url string) (string, error) {
	fmt.Println(tr.Localize(MsgDownloadingPackage(url)))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", tr.WrapError(MsgErrDownloadPackage(url), err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", tr.WrapError(MsgErrDownloadPackage(url), err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return "", tr.WrapError(MsgErrDownloadPackage(url), fmt.Errorf("unexpected status %s", resp.Status))
	}

	f, err := os.CreateTemp("", "release-download-*.tar.gz")
	if err != nil {
		return "", tr.WrapError(MsgErrDownloadPackage(url), err)
	}
	defer func() {
		_ = f.Close()
	}()
	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = os.Remove(f.Name())
		return "", tr.WrapError(MsgErrDownloadPackage(url), err)
	}
	return f.Name(), nil
}
