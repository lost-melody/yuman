package yume

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// runShellScript executes script through /usr/bin/sh, wrapping it in pkexec
// when root privileges are required. pkexecErr localizes a missing pkexec
// binary and runErr localizes a failed command.
func runShellScript(ctx context.Context, script string, usePkexec bool, pkexecErr, runErr *i18n.LocalizeConfig) error {
	argv := []string{"/usr/bin/sh", "-c", script}
	if usePkexec {
		pkexec, err := exec.LookPath("pkexec")
		if err != nil {
			return wrapError(pkexecErr, err)
		}
		argv = append([]string{pkexec, "--keep-cwd"}, argv...)
	}

	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return wrapError(runErr, err)
	}
	return nil
}

// shellQuote quotes a string for safe inclusion in a POSIX shell command.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// shellQuoteAll quotes every string for safe inclusion in a POSIX shell command.
func shellQuoteAll(values []string) []string {
	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = shellQuote(value)
	}
	return quoted
}
