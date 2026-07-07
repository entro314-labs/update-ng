package updater

import "time"

// ScriptStatus represents the state of a script execution
type ScriptStatus int

const (
	StatusPending ScriptStatus = iota
	StatusRunning
	StatusSuccess
	StatusError
	StatusSkipped
)

// ScriptResult holds the result of a script execution
type ScriptResult struct {
	Name        string
	Status      ScriptStatus
	Duration    time.Duration
	Output      string
	Error       error
	Description string
	LastLine    string
}

// ScriptStarted indicates a script has started running
type ScriptStarted struct {
	Name        string
	Description string
}

// ScriptOutput represents streaming output from a running script
type ScriptOutput struct {
	Name   string
	Line   string
	IsLast bool
}

// ScriptCategories maps category names to lists of scripts
var ScriptCategories = map[string][]string{
	"System Package Managers": {
		"update-apt", "update-brew", "update-dnf", "update-macports",
		"update-pacman", "update-zypper", "update-winget", "update-snap",
	},
	"Version Managers": {
		"update-asdf", "update-fnm", "update-gvm", "update-mise", "update-volta",
	},
	"Language Package Managers": {
		"update-npm", "update-pip", "update-cargo", "update-gem", "update-conda",
		"update-mamba", "update-poetry", "update-pdm", "update-rye", "update-uv",
		"update-yarn", "update-bun", "update-deno",
	},
	"Cloud & Infrastructure": {
		"update-aws", "update-gcloud", "update-kubectl", "update-terraform",
		"update-docker", "update-podman",
	},
	"Development Tools": {
		"update-gh", "update-chezmoi", "update-pre-commit", "update-direnv",
	},
	"Shell & Environment": {
		"update-zsh", "update-oh-my-zsh", "update-antidote",
	},
}

// GetScriptDescriptions returns human-readable descriptions for each update script
func GetScriptDescriptions() map[string]string {
	return map[string]string{
		// System Package Managers
		"update-apt":      "APT - Debian/Ubuntu package manager",
		"update-apk":      "APK - Alpine Linux package manager",
		"update-brew":     "Homebrew - macOS package manager",
		"update-choco":    "Chocolatey - Windows package manager",
		"update-dnf":      "DNF - Red Hat/Fedora package manager",
		"update-emerge":   "Emerge - Gentoo package manager",
		"update-eopkg":    "Eopkg - Solus package manager",
		"update-freebsd":  "FreeBSD - FreeBSD system updates",
		"update-guix":     "Guix - GNU Guix package manager",
		"update-macports": "MacPorts - macOS package manager",
		"update-nix":      "Nix - NixOS package manager",
		"update-opkg":     "OPKG - OpenWrt package manager",
		"update-pacman":   "Pacman - Arch Linux package manager",
		"update-pkg":      "PKG - FreeBSD package manager",
		"update-pkg-add":  "PKG_ADD - OpenBSD package manager",
		"update-prt-get":  "PRT-GET - CRUX package manager",
		"update-scoop":    "Scoop - Windows command-line installer",
		"update-slackpkg": "Slackpkg - Slackware package manager",
		"update-snap":     "Snap - Universal Linux packages",
		"update-urpmi":    "URPMI - Mandriva/Mageia package manager",
		"update-winget":   "Windows Package Manager",
		"update-xbps":     "XBPS - Void Linux package manager",
		"update-yay":      "YAY - Arch Linux AUR helper",
		"update-yum":      "YUM - Red Hat/CentOS package manager",
		"update-zypper":   "Zypper - openSUSE package manager",

		// System Updates
		"update-flatpak":        "Flatpak - Linux application distribution",
		"update-mas":            "Mac App Store - macOS app updates",
		"update-macos":          "macOS - System software updates",
		"update-ubuntu-release": "Ubuntu Release - Ubuntu version upgrades",

		// Version Managers
		"update-asdf":   "ASDF - Multi-language version manager",
		"update-fnm":    "Fast Node Manager - Node.js version manager",
		"update-gvm":    "Go Version Manager",
		"update-mise":   "Mise - Modern version manager for tools",
		"update-rustup": "Rustup - Rust toolchain installer",
		"update-volta":  "Volta - JavaScript tool manager",

		// Language Package Managers - JavaScript/TypeScript
		"update-bun":         "Bun - Fast JavaScript runtime and toolkit",
		"update-deno":        "Deno - Secure JavaScript/TypeScript runtime",
		"update-npm":         "NPM - Node.js package manager",
		"update-npm-global":  "NPM Global - Global NPM packages",
		"update-npm-n":       "NPM N - Node.js version manager via NPM",
		"update-pnpm":        "PNPM - Fast, disk space efficient package manager",
		"update-pnpm-global": "PNPM Global - Global PNPM packages",
		"update-yarn":        "Yarn - JavaScript package manager",

		// Language Package Managers - Python
		"update-conda":  "Conda - Data science package manager",
		"update-mamba":  "Mamba - Fast conda-compatible package manager",
		"update-pdm":    "PDM - Modern Python package manager",
		"update-pip":    "Pip - Python package installer",
		"update-pipenv": "Pipenv - Python virtual environment manager",
		"update-poetry": "Poetry - Python dependency manager",
		"update-rye":    "Rye - Experimental Python packaging tool",
		"update-uv":     "UV - Ultra-fast Python package installer",

		// Language Package Managers - Rust
		"update-cargo":                   "Cargo - Rust package manager",
		"update-cargo-projects":          "Cargo Projects - Update multiple Cargo projects",
		"update-cargo-project":           "Cargo Project - Update single Cargo project",
		"update-cargo-project-manifests": "Cargo Project Manifests - Update Cargo.toml files",

		// Language Package Managers - Other
		"update-apm":   "APM - Atom package manager",
		"update-cabal": "Cabal - Haskell package manager",
		"update-gem":   "RubyGems - Ruby package manager",
		"update-mix":   "Mix - Elixir build tool and package manager",
		"update-swift": "Swift - Swift package manager",

		// Mobile Development
		"update-carthage": "Carthage - iOS dependency manager",
		"update-pod":      "CocoaPods - iOS dependency manager",

		// Project Dependencies
		"update-brewfile":            "Brewfile - Homebrew bundle dependencies",
		"update-gemfile":             "Gemfile - Ruby bundle dependencies",
		"update-playwright-via-pnpm": "Playwright via PNPM - Browser automation testing",
		"update-podfile":             "Podfile - iOS CocoaPods dependencies",

		// Cloud & Infrastructure
		"update-aws":       "AWS CLI - Amazon Web Services CLI",
		"update-az":        "Azure CLI - Microsoft Azure CLI",
		"update-docker":    "Docker - Container platform",
		"update-gcloud":    "Google Cloud SDK",
		"update-helm":      "Helm - Kubernetes package manager",
		"update-kubectl":   "Kubectl - Kubernetes command-line tool",
		"update-podman":    "Podman - Alternative container engine to Docker",
		"update-terraform": "Terraform - Infrastructure as Code",

		// Development Tools
		"update-chezmoi":    "Chezmoi - Dotfiles manager",
		"update-direnv":     "Direnv - Environment variable loader",
		"update-gh":         "GitHub CLI",
		"update-pre-commit": "Pre-commit - Git hooks framework",

		// Version Control
		"update-git-pull":           "Git Pull - Update Git repositories",
		"update-git-pull-manifests": "Git Pull Manifests - Update multiple Git repos",
		"update-hg-pull":            "Mercurial Pull - Update Mercurial repositories",
		"update-hg-pull-manifests":  "Mercurial Pull Manifests - Update multiple Mercurial repos",

		// Shell & Environment
		"update-antidote":  "Antidote - Zsh plugin manager",
		"update-oh-my-zsh": "Oh My Zsh - Zsh configuration framework",
		"update-zsh":       "Zsh - Shell updates",

		// Specialized Tools
		"update-cards":  "Cards - Specialized update tool",
		"update-motion": "Motion - Motion detection software updates",
	}
}
