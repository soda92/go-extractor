# Go Extractor

A Go-based, Fyne GUI utility designed for Arch Linux (with Dolphin/KDE integration) that allows you to easily extract archives into custom subfolders.

It replaces user-unfriendly command-line extraction with a simple, keyboard-friendly graphical interface.

---

## ✨ Features

- **One-click / Enter-key Extraction**: Type in your target directory or subfolder and hit `Enter` to start the extraction.
- **KDE Dolphin Integration**: Automatically launches Dolphin to show you the target folder on successful extraction (can be toggled off via the "Open in Dolphin" checkbox).
- **Subfolder Extraction**: Automatically guesses the subfolder name based on the archive name, with an option to toggle subfolder extraction off.
- **Fast and Lightweight**: Uses `7z` under the hood for multi-threaded compression support.

---

## 🛠️ Requirements

Ensure you have the following packages installed:
- `7zip` (for extraction functionality)
- `go` (for compilation)
- `ccache` (optional, to speed up CGO compilation)

---

## 🚀 Local Development

To run the utility locally for testing:
```bash
go run main.go /path/to/archive.zip
```

To compile and install the binary to your local GOPATH/gobin:
```bash
go install
```

---

## 📦 Pacman Packaging (Arch Linux)

This repository includes a `PKGBUILD` and a system-wide Dolphin service menu configuration (`go-extractor.desktop`).

### 1. Build and Install Package
To compile and register the package system-wide with Pacman:
```bash
# Optional: Vendor dependencies to build offline
go mod vendor

# Compile and install
makepkg -si
```

### 2. High-Performance Compiles
To avoid compiling Fyne from scratch on every run, the `PKGBUILD` is optimized to:
- Share your user-level Go build cache (`~/.cache/go-build`) so builds take under **7 seconds**.
- Cache intermediate C compiler steps with `ccache` if installed.

---

## 🌐 Custom Pacman Repository (Self-Hosting)

You can host pre-built binaries on a remote server so other computers can install it without compiling.

### 1. Configure the Remote Server (e.g., Caddy)
Point your web server domain to `/var/www/pacman`. For example, in `/etc/caddy/Caddyfile`:
```caddy
pkg.sodacris.com {
    root * /var/www/pacman
    file_server browse
}
```

### 2. Publish Updates
An automated script, `./publish.sh`, is included in this repository. Running it will automatically compile the package, update the local repository database, and sync it via `rsync`:
```bash
./publish.sh
```

### 3. Add Custom Repo to Clients
On your Arch Linux client machines, append the custom repository to your `/etc/pacman.conf`:
```ini
[custom]
SigLevel = Optional TrustAll
Server = https://pkg.sodacris.com/$arch
```
Update databases and install:
```bash
sudo pacman -Syy go-extractor
```
