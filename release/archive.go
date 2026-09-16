package release

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var (
	MsgErrOpenGzip = func(pkg string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrOpenGzip",
				Other: "open gzip stream from '{{.Package}}'",
			},
			TemplateData: map[string]any{
				"Package": pkg,
			},
		}
	}
	MsgErrArchiveAbsolutePath = func(name string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrArchiveAbsolutePath",
				Other: "archive entry has absolute path: '{{.Name}}'",
			},
			TemplateData: map[string]any{
				"Name": name,
			},
		}
	}
	MsgErrArchiveEscape = func(name string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrArchiveEscape",
				Other: "archive entry escapes destination: '{{.Name}}'",
			},
			TemplateData: map[string]any{
				"Name": name,
			},
		}
	}
)

// ExtractTarGz unpacks a gzipped tarball into destDir, preserving file modes
// and rejecting entries that escape destDir.
func ExtractTarGz(ctx context.Context, pkgPath, destDir string) error {
	f, err := os.Open(pkgPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return tr.WrapError(MsgErrOpenGzip(pkgPath), err)
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
