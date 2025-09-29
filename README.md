# Update Command NG 🚀

**Update Command NG** is a modern, fast system updater written in Go that manages all your package managers, development tools, and system components through a beautiful terminal interface.

## ✨ Features

- 🎨 **Beautiful TUI** - Gorgeous terminal interface with real-time progress
- ⚡ **Parallel Execution** - Update multiple tools simultaneously
- 🎯 **Smart Categorization** - Organized by system, language, cloud tools, etc.
- 🔧 **Selective Updates** - Choose specific tools or categories
- 📊 **Progress Tracking** - Real-time status with completion times
- 🛡️ **Safe & Reliable** - Graceful handling of missing tools
- 🚀 **Single Binary** - No dependencies, just download and run

## 🚀 Quick Start

### Option 1: Download Pre-built Binary (Recommended)
```sh
# macOS/Linux - download from releases
curl -L https://github.com/entro314-labs/update-ng/releases/latest/download/update-ng-$(uname -s)-$(uname -m) -o update-ng
chmod +x update-ng
sudo mv update-ng /usr/local/bin/
```

### Build from Source

```sh
git clone https://github.com/entro314-labs/update-ng.git
cd update-ng
go build -o update-ng main.go
```

## 🎮 Usage

### Basic Usage
```sh
# Update everything with beautiful TUI
update-ng

# Update specific tools only
update-ng --only brew,npm,docker

# Skip certain tools
update-ng --skip macports,conda

# Run sequentially instead of parallel
update-ng --parallel=false

# Disable TUI for scripts/automation
update-ng --tui=false
```

### Advanced Examples

```sh
# Update only Python ecosystem
update-ng --only pip,conda,poetry,uv

# Update only development tools
update-ng --only git,gh,docker,kubectl

# Quick system update
update-ng --only brew,apt,dnf --parallel=true

# Include Rust tools but exclude project scanning
update-ng --only cargo --exclude cargo-projects,cargo-project,cargo-project-manifests

# Skip all Rust project tools specifically
update-ng --skip cargo-project

# Exclude specific tools by exact name
update-ng --exclude update-cargo-projects,update-mas
```

### Filtering Options

Update Command NG provides flexible filtering options to control which tools run:

- **`--only`**: Include only tools matching these patterns (supports partial matching)
- **`--skip`**: Skip tools matching these patterns (supports partial matching)
- **`--exclude`**: Exclude specific tools by exact name (more precise than `--skip`)

**Pattern Matching:**
- Patterns work with or without the `update-` prefix
- `--only cargo` matches `update-cargo`, `update-cargo-projects`, etc.
- `--exclude cargo-projects` matches only `update-cargo-projects`
- Multiple values supported: `--skip conda,macports`

**Order of Operations:**
1. Apply `--only` filter (if specified)
2. Apply `--skip` filter (if specified)
3. Apply `--exclude` filter (if specified)

## 🛠️ Supported Tools

Update Command NG automatically detects and updates the following categories of tools (80+ tools supported):

### System Package Managers

- **APT** (Debian/Ubuntu) - `update-apt`
- **APK** (Alpine Linux) - `update-apk`
- **Chocolatey** (Windows) - `update-choco`
- **DNF** (Red Hat/Fedora) - `update-dnf`
- **Emerge** (Gentoo) - `update-emerge`
- **Eopkg** (Solus) - `update-eopkg`
- **FreeBSD** - `update-freebsd`
- **Guix** (GNU Guix) - `update-guix`
- **Homebrew** (macOS) - `update-brew`
- **MacPorts** (macOS) - `update-macports`
- **Nix** (NixOS) - `update-nix`
- **OPKG** (OpenWrt) - `update-opkg`
- **Pacman** (Arch Linux) - `update-pacman`
- **PKG** (FreeBSD) - `update-pkg`
- **PKG_ADD** (OpenBSD) - `update-pkg-add`
- **PRT-GET** (CRUX) - `update-prt-get`
- **Scoop** (Windows) - `update-scoop`
- **Slackpkg** (Slackware) - `update-slackpkg`
- **Snap** (Universal Linux) - `update-snap`
- **URPMI** (Mandriva/Mageia) - `update-urpmi`
- **Winget** (Windows) - `update-winget`
- **XBPS** (Void Linux) - `update-xbps`
- **YAY** (Arch AUR helper) - `update-yay`
- **YUM** (Red Hat/CentOS) - `update-yum`
- **Zypper** (openSUSE) - `update-zypper`

### System Updates

- **Flatpak** (Linux applications) - `update-flatpak`
- **Mac App Store** (macOS) - `update-mas`
- **macOS System** - `update-macos`
- **Ubuntu Release** - `update-ubuntu-release`

### Version Managers

- **ASDF** (Multi-language) - `update-asdf`
- **FNM** (Node.js) - `update-fnm`
- **GVM** (Go) - `update-gvm`
- **Mise** (Modern tools) - `update-mise`
- **Rustup** (Rust toolchain) - `update-rustup`
- **Volta** (JavaScript) - `update-volta`

### Language Package Managers

#### JavaScript/TypeScript
- **Bun** (Fast runtime & toolkit) - `update-bun`
- **Deno** (Secure runtime) - `update-deno`
- **NPM** (Node.js) - `update-npm`
- **NPM Global** - `update-npm-global`
- **NPM N** (Node version manager) - `update-npm-n`
- **PNPM** (Fast package manager) - `update-pnpm`
- **PNPM Global** - `update-pnpm-global`
- **Yarn** (Package manager) - `update-yarn`

#### Python
- **Conda** (Data science) - `update-conda`
- **Mamba** (Fast conda alternative) - `update-mamba`
- **PDM** (Modern package manager) - `update-pdm`
- **Pip** (Package installer) - `update-pip`
- **Pipenv** (Virtual environments) - `update-pipenv`
- **Poetry** (Dependency manager) - `update-poetry`
- **Rye** (Experimental packaging) - `update-rye`
- **UV** (Ultra-fast installer) - `update-uv`

#### Rust
- **Cargo** (Package manager) - `update-cargo`
- **Cargo Projects** - `update-cargo-projects`
- **Cargo Project** - `update-cargo-project`
- **Cargo Project Manifests** - `update-cargo-project-manifests`

#### Other Languages
- **APM** (Atom packages) - `update-apm`
- **Cabal** (Haskell) - `update-cabal`
- **Gem** (Ruby) - `update-gem`
- **Mix** (Elixir) - `update-mix`
- **Swift** (Swift packages) - `update-swift`

### Mobile Development

- **Carthage** (iOS dependency manager) - `update-carthage`
- **CocoaPods** (iOS) - `update-pod`

### Project Dependencies

- **Brewfile** (Homebrew bundle) - `update-brewfile`
- **Gemfile** (Ruby bundle) - `update-gemfile`
- **Playwright via PNPM** - `update-playwright-via-pnpm`
- **Podfile** (iOS) - `update-podfile`

### Cloud & Infrastructure

- **AWS CLI** - `update-aws`
- **Azure CLI** - `update-az`
- **Docker** (Container platform) - `update-docker`
- **Google Cloud SDK** - `update-gcloud`
- **Helm** (Kubernetes package manager) - `update-helm`
- **Kubectl** (Kubernetes CLI) - `update-kubectl`
- **Podman** (Container alternative) - `update-podman`
- **Terraform** (Infrastructure as Code) - `update-terraform`

### Development Tools

- **Chezmoi** (Dotfiles manager) - `update-chezmoi`
- **Direnv** (Environment loader) - `update-direnv`
- **GitHub CLI** - `update-gh`
- **Pre-commit** (Git hooks) - `update-pre-commit`

### Version Control

- **Git Pull** (Repository updates) - `update-git-pull`
- **Git Pull Manifests** - `update-git-pull-manifests`
- **Mercurial Pull** - `update-hg-pull`
- **Mercurial Pull Manifests** - `update-hg-pull-manifests`

### Shell & Environment

- **Antidote** (Zsh plugin manager) - `update-antidote`
- **Oh My Zsh** (Zsh framework) - `update-oh-my-zsh`
- **Zsh** (Shell updates) - `update-zsh`

### Specialized Tools

- **Cards** - `update-cards`
- **Motion** - `update-motion`

## 📦 Installation

### Download Pre-built Binary (Recommended)

```sh
# macOS/Linux - download from releases
curl -L "https://github.com/entro314-labs/update-ng/releases/download/v1.0.0/update-ng-$(uname -s)-$(uname -m)" -o update-ng
chmod +x update-ng
sudo mv update-ng /usr/local/bin/
```

### Build from Source

```sh
git clone https://github.com/entro314-labs/update-ng.git
cd update-ng
go build -o update-ng main.go
sudo mv update-ng /usr/local/bin/
```

## ⚙️ Configuration

Create a config file at `~/.config/update-ng/config.yaml`:

```yaml
# Default settings
tui: true          # Show beautiful TUI interface
parallel: true     # Run updates in parallel
verbose: true      # Show detailed output
log-file: ""       # Custom log file path

# Default tools to skip
skip:
  - "conda"        # Skip conda if you prefer mamba
  - "macports"     # Skip if you only use Homebrew

# Default tools to run
only: []           # Empty means run all available tools
```

## 🔄 Automation

### Run Daily with Cron

```sh
# Add to your crontab
echo "@daily /usr/local/bin/update-ng --tui=false" | crontab -
```

### Run Weekly with Verbose Output

```sh
# Add to your crontab for detailed logs
echo "@weekly /usr/local/bin/update-ng --verbose=true --tui=false >> /var/log/update-ng.log 2>&1" | crontab -
```

## 🎯 Design Goals

1. **Modern & Fast** - Built with Go for performance and reliability
2. **Beautiful Interface** - Gorgeous TUI with real-time progress and colors
3. **Smart Detection** - Automatically detects installed tools and skips missing ones
4. **Parallel Execution** - Updates multiple tools simultaneously for speed
5. **Developer Friendly** - Easy to extend and customize

## 🤝 Contributing

We welcome contributions! To add support for a new tool:

1. Create a new update script in the `bin/` directory
2. Follow the naming convention: `update-toolname`
3. Add the tool description in `main.go`
4. Test with `update-ng --only toolname`
5. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.
