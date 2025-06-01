# [0.6.0](https://github.com/TecharoHQ/yeet/compare/v0.5.0...v0.6.0) (2025-06-01)


### Bug Fixes

* use sh instead of bash ([#27](https://github.com/TecharoHQ/yeet/issues/27)) ([2e169ef](https://github.com/TecharoHQ/yeet/commit/2e169efafec110d71c8fb596c0a163f902c6f757))


### Features

* support SOURCE_DATE_EPOCH ([#26](https://github.com/TecharoHQ/yeet/issues/26)) ([c01c3db](https://github.com/TecharoHQ/yeet/commit/c01c3db54cddbed5faf1f32a0993620c7aaade9b))
* use mvdan.cc/sh/v3 instead of system sh ([#28](https://github.com/TecharoHQ/yeet/issues/28)) ([5459d92](https://github.com/TecharoHQ/yeet/commit/5459d922e1e3e0aaa3deef6cdf449c4cab3aeeea))

# [0.5.0](https://github.com/TecharoHQ/yeet/compare/v0.4.0...v0.5.0) (2025-05-30)


### Features

* **deb,rpm,tarball:** implement reproducible builds ([49686d8](https://github.com/TecharoHQ/yeet/commit/49686d84f20a6df92378139a6705504621f7c9d9))

# [0.4.0](https://github.com/TecharoHQ/yeet/compare/v0.3.0...v0.4.0) (2025-05-29)


### Features

* **yeetfile:** ppc64le builds ([bbc1e38](https://github.com/TecharoHQ/yeet/commit/bbc1e384f82724660365a4525262180241ec3f06))

# [0.3.0](https://github.com/TecharoHQ/yeet/compare/v0.2.3...v0.3.0) (2025-05-20)


### Features

* **confyg:** export package publicly ([7abba3a](https://github.com/TecharoHQ/yeet/commit/7abba3a1ddcdd9eca4776a80d98851dcfc5005fc))

## [0.2.3](https://github.com/TecharoHQ/yeet/compare/v0.2.2...v0.2.3) (2025-05-09)


### Bug Fixes

* **cmd/yeet:** show build method in --version ([#20](https://github.com/TecharoHQ/yeet/issues/20)) ([ca06ce7](https://github.com/TecharoHQ/yeet/commit/ca06ce7d9247e1d18b8be346e191404e652bd6f9))

## [0.2.2](https://github.com/TecharoHQ/yeet/compare/v0.2.1...v0.2.2) (2025-05-02)


### Bug Fixes

* **mkdeb|mkrpm:** append package name to ${doc} folder ([4976741](https://github.com/TecharoHQ/yeet/commit/4976741c7dba9196d23e25d2fd1ae07af10673e3))

## [0.2.1](https://github.com/TecharoHQ/yeet/compare/v0.2.0...v0.2.1) (2025-04-26)


### Bug Fixes

* make build errors fatal ([#18](https://github.com/TecharoHQ/yeet/issues/18)) ([7a467e7](https://github.com/TecharoHQ/yeet/commit/7a467e7d2b8dc4dd6eb704a9940adb1c9711859e))

# [0.2.0](https://github.com/TecharoHQ/yeet/compare/v0.1.1...v0.2.0) (2025-04-26)


### Bug Fixes

* **internal:** fix version string mangling logic ([#16](https://github.com/TecharoHQ/yeet/issues/16)) ([56e6fa9](https://github.com/TecharoHQ/yeet/commit/56e6fa973d89aa220b0a712c59a751fa8ccfa49c))


### Features

* enforce semver in package versions ([#17](https://github.com/TecharoHQ/yeet/issues/17)) ([178f179](https://github.com/TecharoHQ/yeet/commit/178f17969e17eaf26eb28b9c93a6c24600b5c98c))

# [0.2.0](https://github.com/TecharoHQ/yeet/compare/v0.1.1...v0.2.0) (2025-04-26)


### Bug Fixes

* **internal:** fix version string mangling logic ([#16](https://github.com/TecharoHQ/yeet/issues/16)) ([56e6fa9](https://github.com/TecharoHQ/yeet/commit/56e6fa973d89aa220b0a712c59a751fa8ccfa49c))


### Features

* enforce semver in package versions ([#17](https://github.com/TecharoHQ/yeet/issues/17)) ([178f179](https://github.com/TecharoHQ/yeet/commit/178f17969e17eaf26eb28b9c93a6c24600b5c98c))

## [0.1.1](https://github.com/TecharoHQ/yeet/compare/v0.1.0...v0.1.1) (2025-04-22)

### Bug Fixes

- **internal/mkdeb:** set CGO_ENABLED=0 ([#13](https://github.com/TecharoHQ/yeet/issues/13)) ([5a90b17](https://github.com/TecharoHQ/yeet/commit/5a90b1744ed47e09c6786419f5ecaf172a817606))

# [0.1.0](https://github.com/TecharoHQ/yeet/compare/v0.0.10...v0.1.0) (2025-04-21)

### Features

- **internal:** add --force-git-version flag to override git tag logic ([5f09e47](https://github.com/TecharoHQ/yeet/commit/5f09e4734b838bfcb3ffd99671f6aa280ea81e47))

## [0.0.10](https://github.com/TecharoHQ/yeet/compare/v0.0.9...v0.0.10) (2025-04-21)

### Bug Fixes

- automated release management ([d0efd92](https://github.com/TecharoHQ/yeet/commit/d0efd92f1bb77d2dc8f353dc793c8505e1ee7ddb))
- dispatch releases on main branch ([c1ce6db](https://github.com/TecharoHQ/yeet/commit/c1ce6db03f24e1a8288ae908bd276483933b4327))
- fix release flow? ([d4093e7](https://github.com/TecharoHQ/yeet/commit/d4093e77e7d122f27256b87bdc616884348d0752))
- hack a write token ([d57be0e](https://github.com/TecharoHQ/yeet/commit/d57be0e64ceb6a376578e27421881ae0d0f9e8ed))
- make package builds happen in the release running step ([360e99e](https://github.com/TecharoHQ/yeet/commit/360e99efa745639241806518805c89908e008c11))
- make stable package builds trigger on created ([c4c1955](https://github.com/TecharoHQ/yeet/commit/c4c1955db87004a5e4ab03e2452694439b17a203))

## [0.0.10](https://github.com/TecharoHQ/yeet/compare/v0.0.9...v0.0.10) (2025-04-21)

### Bug Fixes

- automated release management ([d0efd92](https://github.com/TecharoHQ/yeet/commit/d0efd92f1bb77d2dc8f353dc793c8505e1ee7ddb))
- dispatch releases on main branch ([c1ce6db](https://github.com/TecharoHQ/yeet/commit/c1ce6db03f24e1a8288ae908bd276483933b4327))
- fix release flow? ([d4093e7](https://github.com/TecharoHQ/yeet/commit/d4093e77e7d122f27256b87bdc616884348d0752))
- hack a write token ([d57be0e](https://github.com/TecharoHQ/yeet/commit/d57be0e64ceb6a376578e27421881ae0d0f9e8ed))
- make stable package builds trigger on created ([c4c1955](https://github.com/TecharoHQ/yeet/commit/c4c1955db87004a5e4ab03e2452694439b17a203))

## [0.0.10](https://github.com/TecharoHQ/yeet/compare/v0.0.9...v0.0.10) (2025-04-21)

### Bug Fixes

- automated release management ([d0efd92](https://github.com/TecharoHQ/yeet/commit/d0efd92f1bb77d2dc8f353dc793c8505e1ee7ddb))

## v0.0.9

- Enable Gitea package uploading

## v0.0.8

- Add configuration via confyg for package signing
- Added installation instructions to the `README.md`
- Set mtime for deb/rpm package files to unix time 0.

## v0.0.7

Make configuration files for OS packages have mode 0600 by default.

## v0.0.6

- Exit when `--version` is passed.
- Fix CI package autobuilds.

## v0.0.4

Fix go.mod name for project.

## v0.0.3

Fix CI for package builds.

## v0.0.2

- Document package build settings and introduce `yeet.getenv`.

## v0.0.1

- Import source code from [/x/](https://github.com/Xe/x).
