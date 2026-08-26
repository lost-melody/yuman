package yume

import "testing"

func TestParseReleaseAssetName(t *testing.T) {
	tests := []struct {
		name          string
		wantVersion   string
		wantTimestamp string
		wantArch      string
		wantOK        bool
	}{
		{
			name:          "Yume-v3.12.0-20260826130405-linux-x86_64.tar.gz",
			wantVersion:   "v3.12.0",
			wantTimestamp: "20260826130405",
			wantArch:      "x86_64",
			wantOK:        true,
		},
		{
			name:          "Yume-v3.12.0-20260826140000-linux-aarch64.tar.gz",
			wantVersion:   "v3.12.0",
			wantTimestamp: "20260826140000",
			wantArch:      "aarch64",
			wantOK:        true,
		},
		{name: "not-a-tarball.zip", wantOK: false},
		{name: "Other-v1.0-20260826130405-linux-x86_64.tar.gz", wantOK: false},
		{name: "Yume-v1.0-linux-x86_64.tar.gz", wantOK: false},
		{name: "Yume-v1.0-20260826130405-windows-x86_64.tar.gz", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, timestamp, arch, ok := parseReleaseAssetName(tt.name)
			if ok != tt.wantOK {
				t.Fatalf("parseReleaseAssetName(%q) ok = %v, want %v", tt.name, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if version != tt.wantVersion || timestamp != tt.wantTimestamp || arch != tt.wantArch {
				t.Errorf(
					"parseReleaseAssetName(%q) = (%q, %q, %q), want (%q, %q, %q)",
					tt.name, version, timestamp, arch, tt.wantVersion, tt.wantTimestamp, tt.wantArch,
				)
			}
		})
	}
}

func TestSelectReleaseAsset(t *testing.T) {
	assets := []releaseAsset{
		{Name: "Yume-v3.12.0-20260826130405-linux-x86_64.tar.gz", BrowserDownloadURL: "https://example.com/old"},
		{Name: "Yume-v3.12.0-20260826150000-linux-x86_64.tar.gz", BrowserDownloadURL: "https://example.com/new"},
		{Name: "Yume-v3.12.0-20260826150000-linux-aarch64.tar.gz", BrowserDownloadURL: "https://example.com/arm"},
		{Name: "Yume-v3.12.0-20260826150000-linux-x86_64.txt", BrowserDownloadURL: "https://example.com/notgz"},
	}

	got, err := selectReleaseAsset(assets, "x86_64")
	if err != nil {
		t.Fatal(err)
	}
	if want := "https://example.com/new"; got.BrowserDownloadURL != want {
		t.Errorf("selectReleaseAsset() URL = %q, want %q", got.BrowserDownloadURL, want)
	}

	if _, err := selectReleaseAsset(assets, "riscv64"); err == nil {
		t.Error("selectReleaseAsset() = nil error, want no matching asset error")
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"v3.12.0", "v3.10.0", 1},
		{"v3.12.0", "v3.12.0", 0},
		{"3.12.0", "v3.12.0", 0},
		{"v3.9.0", "v3.12.0", -1},
		{"4.0.0", "v3.12.0", 1},
	}
	for _, tt := range tests {
		if got := CompareVersions(tt.a, tt.b); got != tt.want {
			t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
