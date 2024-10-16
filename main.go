package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bitfield/script"
	"github.com/kballard/go-shellquote"
)

var (
	Bubblewrap = "bwrap"
	Melange    = "melange"
	Path       = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
	Term       = "xterm"
	Lang       = "C"
	Home       = "/home/build"
)

func main() {
	prefix, _ := filepath.Abs(".")
	if prefix == "" {
		prefix = "."
	}

	outDir := filepath.Join(prefix, "out")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "(buildroot) %s: %s\n", outDir, err)
	}

	tempRoot := filepath.Join(outDir, "tmp")
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "(buildroot) %s: %s\n", tempRoot, err)
	}

	tempDir, err := os.MkdirTemp(tempRoot, "*")
	if err != nil {
		tempDir = tempRoot
	}

	if err = os.MkdirAll(tempDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "(buildroot) %s: %s\n", tempDir, err)
	}

	melangeCacheDir := filepath.Join(outDir, "cache", "melange")
	if err = os.MkdirAll(melangeCacheDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "(buildroot) %s: %s\n", melangeCacheDir, err)
	}

	apkoCacheDir := filepath.Join(outDir, "cache", "apko")
	if err = os.MkdirAll(apkoCacheDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "(buildroot) %s: %s\n", apkoCacheDir, err)
	}

	cmdline := shellquote.Join(append([]string{
		"bwrap",
		"--clearenv",
		"--setenv", "PATH", Path,
		"--setenv", "TERM", Term,
		"--setenv", "HOME", Home,
		"--setenv", "LC_ALL", Lang,
		"--bind", "/", "/",
		"--bind", tempDir, "/tmp",
		"--dev", "/dev",
		"--proc", "/proc",
		"--",
		"%s",
		"--log-level", "info",
		"--runner", "bubblewrap",
		"--out-dir", outDir,
		"--cache-dir", melangeCacheDir,
		"--apk-cache-dir", apkoCacheDir,
		"--workspace-dir", filepath.Join(tempDir, "workspace"),
		"--guest-dir", filepath.Join(tempDir, "guest"),
	}, os.Args[1:]...)...)

	if !strings.Contains(cmdline, " --src-dir ") {
		cmdline = fmt.Sprintf(cmdline, "%s --empty-workspace")
	}

	if !strings.Contains(cmdline, " --arch ") {
		cmdline = fmt.Sprintf(cmdline, "%s --arch "+runtime.GOARCH)
	}

	if !strings.HasSuffix(cmdline, ".yaml") && !strings.HasSuffix(cmdline, ".yml") {
		cmdline += " " + filepath.Join(prefix, "melange.yaml")
	}

	fmt.Fprintf(os.Stderr, "(buildpkg) %s\n\n", fmt.Sprintf(cmdline, "\nmelange build"))

	cmdline = fmt.Sprintf(cmdline, "melange build")

	_,err = script.Exec(cmdline).Tee().AppendFile("out/buildpkg.log")
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "(buildpkg) error: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprint(os.Stderr, "(buildpkg) done\n")

	script.Exec(shellquote.Join("rm", "-fr", "./out/tmp")).ExitStatus()
}
