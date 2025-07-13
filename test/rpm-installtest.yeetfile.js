const pkgName = rpm.build({
  name: "yeettest",
  description: "Testing RPM installs",
  homepage: "https://xeiaso.net",
  license: "CC0",
  goarch: "all",

  build: ({ doc }) => {
    file.install("./README.md", `${doc}/README.md`);
  },
});

log.println(`export PACKAGE_PATH=${pkgName}`);