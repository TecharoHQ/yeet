# yeet

![enbyware](https://pride-badges.pony.workers.dev/static/v1?label=enbyware&labelColor=%23555&stripeWidth=8&stripeColors=FCF434%2CFFFFFF%2C9C59D1%2C2C2C2C)
![GitHub Issues or Pull Requests by label](https://img.shields.io/github/issues/TecharoHQ/yeet)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/TecharoHQ/yeet)
![language count](https://img.shields.io/github/languages/count/TecharoHQ/yeet)
![repo size](https://img.shields.io/github/repo-size/TecharoHQ/yeet)

Yeet out actions with maximum haste! Declare your build instructions as small JavaScript snippets and let er rip!

For example, here's how you build a Go program into an RPM for x86_64 Linux:

```js
// yeetfile.js
const platform = "linux";
const goarch = "amd64";

rpm.build({
  name: "hello",
  description: "Hello, world!",
  license: "CC0",
  platform,
  goarch,

  build: ({ bin }) => {
    $`go build ./cmd/hello ${bin}/hello`;
  },
});
```

Yeetfiles MUST obey the following rules:

1. Thou shalt never import thine code from another file nor require npm for any reason.
1. If thy task requires common functionality, thou shalt use native interfaces when at all possible.
1. If thy task hath been copied and pasted multiple times, yon task belongeth in a native interface.

See [the API documentation](./doc/api.md) for more information about the exposed API.

## Installation

To install `yeet`, use the following command:

```sh
go install github.com/TecharoHQ/yeet/cmd/yeet@latest
```

## Development

To get started developing for `yeet`, install Go and Node from [Homebrew](https://brew.sh).

```text
brew bundle
npm ci
npm run prepare
```

## Support

For support, please [subscribe to me on Patreon](https://patreon.com/cadey) and ask in the `#yeet` channel in the patron Discord.

## Packaging Status

[![Packaging status](https://repology.org/badge/vertical-allrepos/yeet-js-build-tool.svg?columns=3)](https://repology.org/project/yeet-js-build-tool/versions)

## Contributors

<a href="https://github.com/TecharoHQ/yeet/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=TecharoHQ/yeet" />
</a>

Made with [contrib.rocks](https://contrib.rocks).
