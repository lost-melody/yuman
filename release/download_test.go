package release

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestDownload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("faketarball"))
	}))
	defer server.Close()

	path, err := Download(context.Background(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(path)
	}()

	if !strings.HasSuffix(path, ".tar.gz") {
		t.Errorf("Download() path = %q, want .tar.gz suffix", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "faketarball" {
		t.Errorf("Download() content = %q, want %q", data, "faketarball")
	}
}

func TestDownloadRejectsFailedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer server.Close()

	if _, err := Download(context.Background(), server.URL); err == nil {
		t.Error("Download() = nil error, want unexpected status error")
	}
}
