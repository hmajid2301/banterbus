# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.7.1] - 2025-02-10

### Changed

- Updated all nix dependencies.

## [0.1.7] - 2025-01-07

### Changed

- Updated all nix dependencies.

## [0.1.6] - 2024-12-23

### Changed

- Updated all nix dependencies including playwright to version 1.49.1.

## [0.1.5] - 2024-12-02

### Fixed

- playwright tests running in gitlab ci, include playwright-driver and env variables needed.

## [0.1.4] - 2024-12-01

### Added

- playwright-test so we can use one generic image, see if it makes much difference to job times.

## [0.1.3] - 2024-11-11

### Changed

- Updated go deps and Nix flake.lock file.

## [0.1.2] - 2024-11-01

### Fixed

- Remove playwright-test for now and reduced image size down from 3.2GB -> 1.4GB.

## [0.1.1] - 2024-10-19

### Added

- Updated docker image with gomod2nix and goenv.

## [0.1.0] - 2024-10-05

### Added

- Initial release.
