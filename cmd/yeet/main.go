package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"al.essio.dev/pkg/shellescape"
	yeetver "github.com/Techaro/yeet"
	"github.com/Techaro/yeet/internal/mkdeb"
	"github.com/Techaro/yeet/internal/mkrpm"
	"github.com/Techaro/yeet/internal/mktarball"
	"github.com/Techaro/yeet/internal/pkgmeta"
	"github.com/Techaro/yeet/internal/yeet"
	"github.com/dop251/goja"
)

var (
	fname   = flag.String("fname", "yeetfile.js", "filename for the yeetfile")
	version = flag.Bool("version", false, "if set, print version of yeet and exit")
)

func runcmd(cmdName string, args ...string) string {
	ctx := context.Background()

	slog.Debug("running command", "cmd", cmdName, "args", args)

	result, err := yeet.Output(ctx, cmdName, args...)
	if err != nil {
		panic(err)
	}

	return result
}

func dockerload(fname string) {
	if fname == "" {
		fname = "./result"
	}
	yeet.DockerLoadResult(context.Background(), fname)
}

func dockerbuild(tag string, args ...string) {
	yeet.DockerBuild(context.Background(), yeet.WD, tag, args...)
}

func dockerpush(image string) {
	yeet.DockerPush(context.Background(), image)
}

func buildShellCommand(literals []string, exprs ...any) string {
	var sb strings.Builder
	for i, value := range exprs {
		sb.WriteString(literals[i])
		sb.WriteString(shellescape.Quote(fmt.Sprint(value)))
	}

	sb.WriteString(literals[len(literals)-1])

	return sb.String()
}

func runShellCommand(literals []string, exprs ...any) string {
	shPath, err := exec.LookPath("sh")
	if err != nil {
		panic(err)
	}

	cmd := buildShellCommand(literals, exprs...)

	slog.Debug("running command", "cmd", cmd)
	output, err := yeet.Output(context.Background(), shPath, "-c", cmd)
	if err != nil {
		panic(err)
	}

	return output
}

func hostname() string {
	result, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return result
}

func gitVersion() string {
	vers, err := yeet.GitTag(context.Background())
	if err != nil {
		panic(err)
	}
	return vers[1:]
}

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("yeet version %s\n", yeetver.Version)
	}

	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	defer func() {
		if r := recover(); r != nil {
			slog.Error("error in JS", "err", r)
		}
	}()

	data, err := os.ReadFile(*fname)
	if err != nil {
		log.Fatal(err)
	}

	vm.Set("$", runShellCommand)

	vm.Set("deb", map[string]any{
		"build": func(p pkgmeta.Package) string {
			foutpath, err := mkdeb.Build(p)
			if err != nil {
				panic(err)
			}
			return foutpath
		},
	})

	vm.Set("docker", map[string]any{
		"build": dockerbuild,
		"load":  dockerload,
		"push":  dockerpush,
	})

	vm.Set("file", map[string]any{
		"install": func(src, dst string) {
			if err := mktarball.Copy(src, dst); err != nil {
				panic(err)
			}
		},
	})

	vm.Set("git", map[string]any{
		"repoRoot": func() string {
			return runcmd("git", "rev-parse", "--show-toplevel")
		},
		"tag": gitVersion,
	})

	vm.Set("go", map[string]any{
		"build": func(args ...string) {
			args = append([]string{"build"}, args...)
			runcmd("go", args...)
		},
		"install": func() { runcmd("go", "install") },
	})

	vm.Set("log", map[string]any{
		"println": fmt.Println,
	})

	vm.Set("rpm", map[string]any{
		"build": func(p pkgmeta.Package) string {
			foutpath, err := mkrpm.Build(p)
			if err != nil {
				panic(err)
			}
			return foutpath
		},
	})

	vm.Set("tarball", map[string]any{
		"build": func(p pkgmeta.Package) string {
			foutpath, err := mktarball.Build(p)
			if err != nil {
				panic(err)
			}
			return foutpath
		},
	})

	vm.Set("yeet", map[string]any{
		"cwd":      yeet.WD,
		"datetag":  yeet.DateTag,
		"hostname": hostname(),
		"runcmd":   runcmd,
		"run":      runcmd,
		"setenv":   os.Setenv,
		"goos":     runtime.GOOS,
		"goarch":   runtime.GOARCH,
	})

	if _, err := vm.RunScript(*fname, string(data)); err != nil {
		fmt.Fprintf(os.Stderr, "error running %s: %v", *fname, err)
		os.Exit(1)
	}
}
