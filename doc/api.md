# Yeetfile API

Yeet uses [goja](https://pkg.go.dev/github.com/dop251/goja#section-readme) to execute JavaScript. As such, it does not have access to NPM or other external JavaScript libraries. You also cannot import code/data from other files. These are not planned for inclusion into yeet. If functionality is required, it should be added to yeet itself.

To make it useful, yeet exposes a bunch of helper objects full of tools. These tools fall in a few categories, each has its own section.

## `$`

`$` lets you construct shell commands using [tagged templates](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Template_literals). This lets you build whatever shell commands you want by mixing Go and JavaScript values freely.

Example:

```js
$`CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -s -w -extldflags "-static" -X "within.website/x.Version=${git.tag()}"`;
```

## `deb`

Helpers for building Debian packages.

### `deb.build`

Builds a Debian package with a descriptor object. See the native packages section for more information. The important part of this is your `build` function. The `build` function is what will turn your package source code into an executable in `out` somehow.

The resulting Debian package path will be returned as a string.

Usage:

`deb.build(package);`

```js
["amd64", "arm64"].forEach((goarch) =>
  deb.build({
    name: "yeet",
    description: "Yeet out actions with maximum haste!",
    homepage: "https://techaro.lol",
    license: "MIT",
    goarch,

    build: ({ bin }) => {
      go.build("-o", `${bin}/yeet`, "./cmd/yeet");
    },
  }),
);
```

## `docker`

Aliases for `docker` commands.

### `docker.build`

An alias for the `docker build` command. Builds a docker image in the current working directory's Dockerfile.

Usage:

`docker.build(tag);`

```js
docker.build("ghcr.io/xe/site/bin");
docker.push("ghcr.io/xe/site/bin");
```

### `docker.push`

Pushes a docker image to a registry. Analogous to `docker push` in the CLI.

Usage:

`docker.push(tag);`

```js
docker.build("ghcr.io/xe/site/bin");
docker.push("ghcr.io/xe/site/bin");
```

## `file`

### `file.install`

Copies from a file from one place to another whilst preserving the file mode, analogous to `install -d` on Linux. Automatically creates directories in the `dest` path if they don't exist already.

Usage:

`file.install(src, dest);`

```js
file.install("LICENSE", `${doc}/LICENSE`);
```

## `git`

Helpers for the Git version control system.

### `git.repoRoot`

Returns the repository root as a string.

`git.repoRoot();`

```js
const repoRoot = git.repoRoot();

file.copy(`${repoRoot}/LICENSE`, `${doc}/LICENSE`);
```

### `git.tag`

Returns the output of `git describe --tags`. Useful for getting the "current version" of the repo, where the current version will likely be different forward in time than it is backwards in time.

Usage:

`git.tag();`

```js
const version = git.tag();
```

## `gitea`

Helpers for integrating with Gitea servers.

### `gitea.uploadPackage`

Uploads a binary package to Gitea, silently failing if the package is not a `.deb` or `.rpm` file. Gitea configuration is done with flags or the configuration file.

Usage:

`gitea.uploadPackage(owner, distro, component, fname)`

```js
gitea.uploadPackage(
  "Techaro",
  "yeet",
  "unstable",
  "./var/yeet-0.0.8.x86_64.rpm",
);
```

## `go`

Helpers for the Go programming language.

### `go.build`

Runs `go build` in the current working directory with any extra arguments passed in. This is useful for building and installing Go programs in an RPM build context.

Usage:

`go.build(args);`

```js
go.build("-o", `${out}/usr/bin/`);
```

### `go.install`

Runs `go install`. Not useful for cross-compilation.

Usage:

`go.install();`

```js
go.install();
```

## `log`

Logging functions.

### `log.println`

Prints log data to standard output.

Usage:

`log.println(...);`

```js
log.println(`built package ${pkgPath}`);
```

## `rpm`

Helpers for building RPM packages and docker images out of a constellation of RPM packages.

### `rpm.build`

Builds an RPM package with a descriptor object. See the RPM packages section for more information. The important part of this is your `build` function. The `build` function is what will turn your package source code into an executable in `out` somehow. Everything in `out` corresponds 1:1 with paths in the resulting RPM.

The resulting RPM path will be returned as a string.

Usage:

`rpm.build(package);`

```js
["amd64", "arm64"].forEach((goarch) =>
  rpm.build({
    name: "yeet",
    description: "Yeet out actions with maximum haste!",
    homepage: "https://techaro.lol",
    license: "MIT",
    goarch,

    build: ({ bin }) => {
      go.build("-o", `${bin}/yeet`, "./cmd/yeet");
    },
  }),
);
```

## `yeet`

This contains various "other" functions that don't have a good place to put them.

### `yeet.cwd`

The current working directory. This is a constant value and is not updated at runtime.

Usage:

```js
log.println(yeet.cwd);
```

### `yeet.dateTag`

A constant string representing the time that yeet was started in UTC. It is formatted in terms of `YYYYmmDDhhMM`. This is not updated at runtime. You can use it for a "unique" value per invocation of yeet (assuming you aren't a time traveler).

Usage:

```js
docker.build(`ghcr.io/xe/site/bin:${git.tag()}-${yeet.dateTag}`);
```

### `yeet.getenv`

Gets an environment variable and returns it as a string, optionally returning an empty string if the variable is not found.

Usage:

`yeet.getenv(name);`

```js
const someValue = yeet.getenv("SOME_VALUE");
```

### `yeet.run` / `yeet.runcmd`

Runs an arbitrary command and returns any output as a string.

Usage:

`yeet.run(cmd, arg1, arg2, ...);`

```js
yeet.run(
  "protoc",
  "--proto-path=.",
  `--proto-path=${git.repoRoot()}/proto`,
  "foo.proto",
);
```

### `yeet.setenv`

Sets an environment variable for the process yeet is running in and all children.

Usage:

`yeet.setenv(key, val);`

```js
yeet.setenv("GOOS", "linux");
```

### `yeet.goos` / `yeet.goarch`

The GOOS/GOARCH value that yeet was built for. This typically corresponds with the OS and CPU architecture that yeet is running on.

## Building native packages

When using the `deb.build`, `rpm.build`, or `tarball.build` functions, you can create native packages from arbitrary yeet expressions. This allows you to cross-compile native packages from a macOS or other Linux system. As an example, here is how the yeet packages are built:

```js
["amd64", "arm64"].forEach((goarch) =>
  [deb, rpm, tarball].forEach((method) =>
    method.build({
      name: "yeet",
      description: "Yeet out scripts with maximum haste!",
      homepage: "https://techaro.lol",
      license: "MIT",
      goarch,

      build: ({ bin }) => {
        go.build("-o", `${bin}/yeet`, "./cmd/yeet");
      },
    }),
  ),
);
```

### Build settings

The following settings are supported:

| Name            | Example                                    | Description                                                                                                                                                                                                                    |
| :-------------- | :----------------------------------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`          | `xeiaso.net-yeet`                          | The name of the package. This should be unique across the system.                                                                                                                                                              |
| `version`       | `1.0.0`                                    | The version of the package, if not set then it will be inferred from the git version.                                                                                                                                          |
| `description`   | `Yeet out scripts with haste!`             | The human-readable description of the package.                                                                                                                                                                                 |
| `homepage`      | `https://xeiaso.net`                       | The URL for the homepage of the package.                                                                                                                                                                                       |
| `group`         | `Network`                                  | If set, the RPM group that this package belongs to.                                                                                                                                                                            |
| `license`       | `MIT`                                      | The license that the contents of this package is under.                                                                                                                                                                        |
| `goarch`        | `amd64` / `arm64`                          | The GOARCH value corresponding to the architecture that the RPM is being built for. If you want to build a `noarch` package, put `any` here.                                                                                   |
| `replaces`      | `["foo", "bar"]`                           | Any packages that this package conflicts with or replaces.                                                                                                                                                                     |
| `depends`       | `["foo", "bar"]`                           | Any packages that this package depends on (such as C libraries for CGo code).                                                                                                                                                  |
| `emptyDirs`     | `["/var/lib/yeet"]`                        | Any empty directories that should be created when the package is installed.                                                                                                                                                    |
| `configFiles`   | `{"./.env.example": "/var/lib/yeet/.env"}` | Any configuration files that should be copied over on install, but managed by administrators after installation.                                                                                                               |
| `documentation` | `{"./README.md": "README.md"}`             | Any documentation files that should be copied to the `doc` folder of a tarball or be put in `/usr/share/doc` in an OS package. Try to include enough documentation that users can troubleshoot the program completely offline. |
| `files`         | `{}`                                       | Any other static files that should be copied in-place to a path in the target filesystem.                                                                                                                                      |

Packages MUST define a `build` function and tarball packages MAY define a `mkFilename` function.

### `build` function

Every package definition MUST contain a `build` function that describes how to build the software. The build function takes one argument and returns nothing. If the build fails, throw an Exception with `throw`.

The signature of `build` roughly follows this TypeScript type:

```ts
interface BuildInput {
  // output folder, usually the package root
  out: string;
  // binary folder, ${out}/bin for tarballs or ${out}/usr/bin for OS packages
  bin: string;
  // documentation folder, ${out}/doc for tarballs or ${out}/usr/share/${pkg.name}/doc for OS packages
  doc: string;
  // configuration folder, ${out}/run for tarballs or ${out}/etc/${pkg.name} for OS packages
  etc: string;
  // systemd unit folder, ${out}/run for tarballs or ${out}/usr/lib/systemd/system for OS packages
  systemd: string;
}

function build({...}: BuildInput) => {
  // ...
};
```

### `mkFilename` function

When building a tarball, you MAY define a `mkFilename` function to customize the generated filename. If no `mkFilename` function is specified, the filename defaults to:

```js
const mkFilename = ({name, version, platform, goarch}) =
  `${name}-${version}-${platform}-${goarch}`;
```

For example, to reduce the filename to the name and the version:

```js
tarball.build({
  // ...
  mkFilename: ({ name, version }) => `${name}-${version}`,
  // ...
});
```
