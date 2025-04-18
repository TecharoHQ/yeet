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

## Support

For support, please [subscribe to me on Patreon](https://patreon.com/cadey) and ask in the `#yeet` channel in the patron Discord.
