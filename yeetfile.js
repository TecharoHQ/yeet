const pkgs = [];

[
  "amd64",
  "arm64",
  "ppc64le",
].forEach((goarch) =>
  [deb, rpm, tarball].forEach((method) => {
    pkgs.push(
      method.build({
        name: "yeet",
        description: "Yeet out scripts with maximum haste!",
        homepage: "https://techaro.lol",
        license: "MIT",
        goarch,

        documentation: {
          "README.md": "README.md",
          "doc/api.md": "api.md",
        },

        build: ({ bin }) => {
          $`go build -o ${bin}/yeet -ldflags '-s -w -extldflags "-static" -X "github.com/TecharoHQ/yeet.Version=${git.tag()}" -X "github.com/TecharoHQ/yeet.BuildMethod=${method.name}"' ./cmd/yeet`;
        },
      }),
    );
  }),
);

tarball.build({
  name: "yeet-src-vendor",
  license: "MIT",
  // XXX(Xe): This is needed otherwise go will be very sad.
  platform: yeet.goos,
  goarch: yeet.goarch,

  build: ({ out }) => {
    // prepare clean checkout in $out
    $`git archive --format=tar HEAD | tar xC ${out}`;
    // vendor Go dependencies
    $`cd ${out} && go mod vendor`;
    // write VERSION file
    $`echo ${git.tag()} > ${out}/VERSION`;
  },

  mkFilename: ({ name, version }) => `${name}-${version}`,
});