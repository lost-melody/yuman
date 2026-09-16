package release

import "testing"

func TestAPIURL(t *testing.T) {
	got := APIURL("forfudan/yumete")
	want := "https://api.github.com/repos/forfudan/yumete/releases/latest"
	if got != want {
		t.Errorf("APIURL() = %q, want %q", got, want)
	}
}

func TestMachineArch(t *testing.T) {
	tests := []struct {
		goarch string
		want   string
	}{
		{"amd64", "x86_64"},
		{"arm64", "aarch64"},
		{"386", "i686"},
		{"loong64", "loongarch64"},
		{"riscv64", "riscv64"},
	}
	for _, tt := range tests {
		if got := MachineArch(tt.goarch); got != tt.want {
			t.Errorf("MachineArch(%q) = %q, want %q", tt.goarch, got, tt.want)
		}
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

func TestFormatSize(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{0, "0 B"},
		{-1, "0 B"},
		{512, "512 B"},
		{2048, "2.00 KiB"},
		{1048576, "1.00 MiB"},
	}
	for _, tt := range tests {
		if got := FormatSize(tt.size); got != tt.want {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.size, got, tt.want)
		}
	}
}
