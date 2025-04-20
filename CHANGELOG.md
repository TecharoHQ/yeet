# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
