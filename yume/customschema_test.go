package yume

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDivisionPaths(t *testing.T) {
	t.Run("empty uses temp output", func(t *testing.T) {
		in, out, cleanup, err := divisionPaths("")
		if err != nil {
			t.Fatal(err)
		}
		if in != "/dev/null" {
			t.Errorf("input = %q, want %q", in, "/dev/null")
		}
		if filepath.Base(out) != "division.ydiv" {
			t.Errorf("output = %q, want a path ending in division.ydiv", out)
		}
		outDir := filepath.Dir(out)
		if _, err := os.Stat(outDir); err != nil {
			t.Errorf("output dir should exist: %v", err)
		}
		cleanup()
		if _, err := os.Stat(outDir); !os.IsNotExist(err) {
			t.Errorf("cleanup should remove the output dir, stat err = %v", err)
		}
	})

	t.Run("txt extension replaced", func(t *testing.T) {
		in, out, cleanup, err := divisionPaths("/tmp/foo.txt")
		if err != nil {
			t.Fatal(err)
		}
		cleanup()
		if in != "/tmp/foo.txt" {
			t.Errorf("input = %q, want %q", in, "/tmp/foo.txt")
		}
		if out != "/tmp/foo.ydiv" {
			t.Errorf("output = %q, want %q", out, "/tmp/foo.ydiv")
		}
	})

	t.Run("non-txt extension replaced", func(t *testing.T) {
		in, out, cleanup, err := divisionPaths("/tmp/foo.div")
		if err != nil {
			t.Fatal(err)
		}
		cleanup()
		if in != "/tmp/foo.div" {
			t.Errorf("input = %q, want %q", in, "/tmp/foo.div")
		}
		if out != "/tmp/foo.ydiv" {
			t.Errorf("output = %q, want %q", out, "/tmp/foo.ydiv")
		}
	})
}

func TestResolveCustomRoot(t *testing.T) {
	t.Run("XDG_DATA_HOME set", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "/custom/data")
		got, err := resolveCustomRoot()
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join("/custom/data", "yume", "data", "custom")
		if got != want {
			t.Errorf("resolveCustomRoot() = %q, want %q", got, want)
		}
	})

	t.Run("XDG_DATA_HOME fallback", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "")
		t.Setenv("HOME", "/home/test")
		got, err := resolveCustomRoot()
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join("/home/test", ".local", "share", "yume", "data", "custom")
		if got != want {
			t.Errorf("resolveCustomRoot() = %q, want %q", got, want)
		}
	})
}

func TestResolveCompileBinary(t *testing.T) {
	// Prevent the $HOME/.local/bin fallback from matching the developer's own
	// yume-compile during tests.
	t.Setenv("HOME", t.TempDir())

	writeExec := func(t *testing.T, path string) {
		t.Helper()
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("prefers PATH", func(t *testing.T) {
		dir := t.TempDir()
		bin := filepath.Join(dir, compileBinaryName)
		writeExec(t, bin)
		t.Setenv("PATH", dir)

		got, err := resolveCompileBinary()
		if err != nil {
			t.Fatal(err)
		}
		if got != bin {
			t.Errorf("resolveCompileBinary() = %q, want %q", got, bin)
		}
	})

	t.Run("falls back to executable dir", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())

		dir := t.TempDir()
		bin := filepath.Join(dir, compileBinaryName)
		writeExec(t, bin)
		prev := executablePath
		executablePath = func() (string, error) { return filepath.Join(dir, "yuman"), nil }
		t.Cleanup(func() { executablePath = prev })

		got, err := resolveCompileBinary()
		if err != nil {
			t.Fatal(err)
		}
		if got != bin {
			t.Errorf("resolveCompileBinary() = %q, want %q", got, bin)
		}
	})

	t.Run("follows symlinked executable", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())

		realDir := t.TempDir()
		realExe := filepath.Join(realDir, "yuman")
		writeExec(t, realExe)
		bin := filepath.Join(realDir, compileBinaryName)
		writeExec(t, bin)

		link := filepath.Join(t.TempDir(), "yuman")
		if err := os.Symlink(realExe, link); err != nil {
			t.Fatal(err)
		}
		prev := executablePath
		executablePath = func() (string, error) { return link, nil }
		t.Cleanup(func() { executablePath = prev })

		got, err := resolveCompileBinary()
		if err != nil {
			t.Fatal(err)
		}
		if got != bin {
			t.Errorf("resolveCompileBinary() = %q, want %q", got, bin)
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())

		prev := executablePath
		executablePath = func() (string, error) { return filepath.Join(t.TempDir(), "yuman"), nil }
		t.Cleanup(func() { executablePath = prev })

		if _, err := resolveCompileBinary(); !errors.Is(err, exec.ErrNotFound) {
			t.Errorf("resolveCompileBinary() err = %v, want exec.ErrNotFound", err)
		}
	})
}

func TestCompileQuestions(t *testing.T) {
	stdout := "  ? name Your name?\nplain line\n?age How old are you?\n?single\n"
	got := compileQuestions(stdout)
	want := []CompileQuestion{
		{Tag: "name", Prompt: "Your name?"},
		{Tag: "age", Prompt: "How old are you?"},
		{Tag: "single", Prompt: ""},
	}
	if len(got) != len(want) {
		t.Fatalf("compileQuestions() = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("compileQuestions()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestCompileAnswerArgs(t *testing.T) {
	questions := []CompileQuestion{
		{Tag: "name", Prompt: "Your name?"},
		{Tag: "age", Prompt: "How old are you?"},
	}
	answers := []string{"alice", "30"}

	got := compileAnswerArgs(questions, answers)
	want := []string{"--answers", "name=alice", "--answers", "age=30"}
	if len(got) != len(want) {
		t.Fatalf("compileAnswerArgs() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("compileAnswerArgs()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParseSlotList(t *testing.T) {
	out := "42\tschema-tag\tfoo\t3\t;'\t1\t0\t1\n7\tanother-tag\tbar\t4\t;'\t1\t1\t0\n"

	schemas := parseSlotList(out)
	if len(schemas) != 2 {
		t.Fatalf("parseSlotList() = %d schemas, want 2", len(schemas))
	}
	if schemas[1].ID != "7" || schemas[1].Name != "bar" {
		t.Errorf("schemas[1] = %+v, want id=7 name=bar", schemas[1])
	}
	if schemas[0].MaxCodeLength != "3" {
		t.Errorf("schemas[0].MaxCodeLength = %q, want %q", schemas[0].MaxCodeLength, "3")
	}
}

func TestRunCompile(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-compile")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho '? ask name'\necho 'slot-output'\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, err := runCompile(context.Background(), script, false, "--slot-create", "/root")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "? ask name") || !strings.Contains(out, "slot-output") {
		t.Errorf("runCompile() stdout = %q, want the script output", out)
	}
}
