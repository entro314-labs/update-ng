# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- GoReleaser infrastructure for automated releases
- Support for Homebrew tap installation
- Support for AUR (Arch User Repository) installation
- Support for Winget (Windows Package Manager)
- Docker images for containerized usage
- Multi-platform binaries (Linux, macOS, Windows, BSD variants)
- Multi-architecture support (amd64, arm64, arm, 386)
- Automated changelog generation
- GPG signing support for releases
- SBOM (Software Bill of Materials) generation

## [1.0.3] - 2025-09-24

### Changed
- Improved update scripts and error handling
- Enhanced verbose output and logging
- Bumped version metadata

## [1.0.2] - Previous release

### Added
- Advanced script filtering capabilities
- Error categorization improvements

### Changed
- Revamped README documentation
- Expanded script descriptions

## [1.0.0] - Initial release

### Added
- Beautiful TUI with real-time progress tracking
- Parallel and sequential execution modes
- Support for 80+ package managers and tools
- Selective update filtering
- Smart tool detection and graceful skipping
- Comprehensive logging system
- Configuration file support
- CI/CD workflows

### Features
- System Package Managers (APT, Homebrew, Pacman, etc.)
- Version Managers (ASDF, Mise, Rustup, etc.)
- Language Package Managers (npm, pip, cargo, etc.)
- Cloud & Infrastructure tools (AWS, Docker, Kubectl, etc.)
- Development tools (GitHub CLI, pre-commit, etc.)
- Shell & Environment tools (Zsh, Oh My Zsh, etc.)

[Unreleased]: https://github.com/entro314-labs/update-ng/compare/v1.0.3...HEAD
[1.0.3]: https://github.com/entro314-labs/update-ng/releases/tag/v1.0.3
[1.0.2]: https://github.com/entro314-labs/update-ng/releases/tag/v1.0.2
[1.0.0]: https://github.com/entro314-labs/update-ng/releases/tag/v1.0.0
