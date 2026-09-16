package release

import (
	"fmt"
	"runtime"
	"strings"
)

// HostArch is the local machine architecture, expressed as a runtime.GOARCH
// value. It is a variable so tests can override it.
var HostArch = runtime.GOARCH

// MachineArch maps a Go architecture name (runtime.GOARCH) to the uname -m
// style used by release asset names.
func MachineArch(goarch string) string {
	switch goarch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	case "386":
		return "i686"
	case "loong64":
		return "loongarch64"
	default:
		return goarch
	}
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
