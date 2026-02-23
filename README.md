# Go Linux Cleaner

A safe, performant, and feature-rich Linux system cleaner written in Go. featuring a modern TUI (Text User Interface) and a scriptable CLI mode.

## Features

- **Interactive TUI**: Visual dashboard to select and run cleaners.
- **APT Cache**: Cleans `/var/cache/apt/archives`.
- **System Logs**: Vacuums `journald` and removes old log files.
- **Trash**: Empties user trash (`~/.local/share/Trash`).
- **Caches**:
  - Thumbnails (`~/.cache/thumbnails`)
  - Browsers (Chrome, Firefox, Brave, etc.)
  - Go Build Cache
  - Generic Cache Scanner
- **Docker**: Prunes unused system objects.
- **Large Files**: Interactive scanner for old (>30 days), large (>100MB) files.

## Installation

```bash
git clone https://github.com/yourusername/linux-cleaner.git
cd linux-cleaner
go build -ldflags "-s -w" -o goclean ./cmd/goclean
```

## Usage

### Interactive TUI (Default)
Simply run the binary to launch the TUI:
```bash
./goclean
```
- Use **Up/Down** keys to navigate.
- Use **Space** to toggle selection.
- Press **Enter** to clean selected items.

### CLI Mode (Scriptable)
Use `--no-tui` for standard command-line output.

```bash
# Dry run (simulate without deleting)
./goclean --no-tui --dry-run

# Run non-interactively (skip confirmations)
sudo ./goclean --no-tui --yes
```

> **Note**: Some cleaners (APT, Docker, Logs) may require `sudo` privileges.

## Safety
- **Dry Run**: Always verify with `--dry-run` first.
- **Confirmations**: The tool asks for confirmation before deleting large files or running mass cleanups (unless `--yes` is used).
- **Exclusions**: Hidden folders are skipped during large file scans to protect config files.

## License
MIT
