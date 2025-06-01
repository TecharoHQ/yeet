package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"al.essio.dev/pkg/shellescape"
	yeetver "github.com/TecharoHQ/yeet"
	"github.com/TecharoHQ/yeet/confyg/flagconfyg"
	"github.com/TecharoHQ/yeet/internal/gitea"
	"github.com/TecharoHQ/yeet/internal/mkdeb"
	"github.com/TecharoHQ/yeet/internal/mkrpm"
	"github.com/TecharoHQ/yeet/internal/mktarball"
	"github.com/TecharoHQ/yeet/internal/pkgmeta"
	"github.com/TecharoHQ/yeet/internal/yeet"
	"github.com/dop251/goja"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

var (
	config  = flag.String("config", configFileLocation(), "configuration file, if set (see flagconfyg(4))")
	fname   = flag.String("fname", "yeetfile.js", "filename for the yeetfile")
	version = flag.Bool("version", false, "if set, print version of yeet and exit")
)

func configFileLocation() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		//ln.Error(context.Background(), err, ln.Debug("can't read config dir"))
		return ""
	}

	dir = filepath.Join(dir, "techaro.lol", "yeet")
	os.MkdirAll(dir, 0700)

	return filepath.Join(dir, filepath.Base(os.Args[0])+".config")
}

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
	fmt.Fprintln(&sb, "set -e")

	for i, value := range exprs {
		sb.WriteString(literals[i])
		sb.WriteString(shellescape.Quote(fmt.Sprint(value)))
	}

	sb.WriteString(literals[len(literals)-1])

	return sb.String()
}

func runShellCommand(ctx context.Context, literals []string, exprs ...any) (string, error) {
	src := buildShellCommand(literals, exprs...)

	slog.Debug("running command", "src", src)

	file, err := syntax.NewParser().Parse(strings.NewReader(src), "")
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	runner, err := interp.New(
		interp.StdIO(nil, &buf, os.Stderr),
	)
	if err != nil {
		return "", err
	}

	if err := runner.Run(ctx, file); err != nil {
		return "", err
	}

	slog.Debug("command output", "src", src, "output", buf.String())

	return buf.String(), nil
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
	return vers
}

func main() {
	flag.Parse()
	ctx := context.Background()

	if *config != "" {
		flagconfyg.CmdParse(ctx, *config)
	}
	flag.Parse()

	if *version {
		fmt.Printf("yeet version %s, built via %s\n", yeetver.Version, yeetver.BuildMethod)
		return
	}

	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	defer func() {
		if r := recover(); r != nil {
			slog.Error("error in JS", "err", r)
			os.Exit(1)
		}
	}()

	data, err := os.ReadFile(*fname)
	if err != nil {
		log.Fatal(err)
	}

	vm.Set("$", func(literals []string, exprs ...any) string {
		result, err := runShellCommand(ctx, literals, exprs...)
		if err != nil {
			panic(err)
		}
		return result
	})

	vm.Set("deb", map[string]any{
		"build": func(p pkgmeta.Package) string {
			foutpath, err := mkdeb.Build(p)
			if err != nil {
				panic(err)
			}
			return foutpath
		},
		"name": "debian",
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

	vm.Set("gitea", map[string]any{
		"uploadPackage": func(owner, distro, component, fname string) {
			if err := gitea.UploadPackage(ctx, http.DefaultClient, owner, distro, component, fname); err != nil {
				panic(err)
			}
		},
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
		"name": "rpm",
	})

	vm.Set("tarball", map[string]any{
		"build": func(p pkgmeta.Package) string {
			foutpath, err := mktarball.Build(p)
			if err != nil {
				panic(err)
			}
			return foutpath
		},
		"name": "tarball",
	})

	vm.Set("yeet", map[string]any{
		"cwd":      yeet.WD,
		"datetag":  yeet.DateTag,
		"hostname": hostname(),
		"runcmd":   runcmd,
		"run":      runcmd,
		"setenv":   os.Setenv,
		"getenv":   os.Getenv,
		"goos":     runtime.GOOS,
		"goarch":   runtime.GOARCH,
	})

	if _, err := vm.RunScript(*fname, string(data)); err != nil {
		fmt.Fprintf(os.Stderr, "error running %s: %v", *fname, err)
		os.Exit(1)
	}
}
