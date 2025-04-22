## [0.1.1](https://github.com/TecharoHQ/yeet/compare/v0.1.0...v0.1.1) (2025-04-22)


### Bug Fixes

* **internal/mkdeb:** set CGO_ENABLED=0 ([#13](https://github.com/TecharoHQ/yeet/issues/13)) ([5a90b17](https://github.com/TecharoHQ/yeet/commit/5a90b1744ed47e09c6786419f5ecaf172a817606))

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
