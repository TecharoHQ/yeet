## [0.0.10](https://github.com/TecharoHQ/yeet/compare/v0.0.9...v0.0.10) (2025-04-21)


### Bug Fixes

* automated release management ([d0efd92](https://github.com/TecharoHQ/yeet/commit/d0efd92f1bb77d2dc8f353dc793c8505e1ee7ddb))

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
