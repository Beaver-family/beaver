<p align="center">
  <img src="https://stealth.blr1.cdn.digitaloceanspaces.com/beaver.png" width="160" alt="Beaver logo" />
</p>

<h1 align="center">Beaver</h1>

<p align="center">
  A fast, keyboard-driven terminal file manager built with Go.
</p>

<p align="center">
  <a href="https://github.com/Beaver-family/beaver/releases/latest"><img src="https://img.shields.io/github/v/release/Beaver-family/beaver?color=5DCAA5&label=latest" alt="Latest Release" /></a>
  <a href="https://github.com/Beaver-family/beaver/blob/main/LICENSE"><img src="https://img.shields.io/github/license/Beaver-family/beaver?color=534AB7" alt="License" /></a>
  <img src="https://img.shields.io/badge/built%20with-Go-00ADD8?logo=go" alt="Built with Go" />
</p>

---

## Features

- **File tree** — navigate your filesystem with an expandable tree, folders-first sorting, hidden files skipped
- **Git status colors** — modified files are yellow, staged files are green, untracked files are red; colors propagate up to parent directories
- **Syntax-highlighted preview** — 300+ languages via [Chroma](https://github.com/alecthomas/chroma), Dracula theme
- **In-app editor** — edit files without leaving the terminal, with line numbers and nano-style save/discard prompts
- **Full file operations** — delete, rename, new file, new folder, copy, cut, paste with confirmation dialogs
- **Fuzzy file search** — instantly filter files across the entire tree by name
- **Content grep** — search inside files with async line-by-line results
- **Claude AI agent** — press `c` to open an embedded AI assistant that can read, write, and create files in your project; choose from Opus, Sonnet, or Haiku
- **Clipboard indicator** — cut items turn red, copied items turn teal in the tree
- **Live clock** — date and time always visible in the status bar
- **Zero config** — works out of the box, optional `config.yaml` for customisation

---

## Installation

### Homebrew (macOS / Linux)

```sh
brew tap Beaver-family/tap
brew install beaver
```

### Go

```sh
go install github.com/Beaver-family/beaver/cmd/beaver@latest
```

### Binary

Download the pre-built binary for your platform from the [releases page](https://github.com/Beaver-family/beaver/releases/latest):

| Platform | File |
|---|---|
| macOS (Apple Silicon) | `beaver_*_darwin_arm64.tar.gz` |
| macOS (Intel) | `beaver_*_darwin_amd64.tar.gz` |
| Linux (x86_64) | `beaver_*_linux_amd64.tar.gz` |
| Linux (ARM64) | `beaver_*_linux_arm64.tar.gz` |
| Windows | `beaver_*_windows_amd64.zip` |

Extract and place the `beaver` binary somewhere on your `$PATH`.

---

## Usage

```sh
beaver          # open in current directory
beaver --version
```

Beaver opens in whatever directory you run it from.

---

## Keybindings

### Navigation

| Key | Action |
|---|---|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `enter` / `→` | Open file / expand folder |
| `tab` | Switch to preview panel |
| `esc` / `←` | Back to file tree |
| `q` | Quit |

### Preview Panel

| Key | Action |
|---|---|
| `↑` / `↓` | Scroll line |
| `ctrl+d` / `ctrl+u` | Half page down / up |
| `pgdn` / `pgup` | Full page |
| `g` / `G` | Top / bottom |
| `e` | Edit file |

### File Operations

| Key | Action |
|---|---|
| `d` | Delete (asks confirmation) |
| `r` | Rename |
| `n` | New file |
| `N` | New folder |
| `y` | Copy to clipboard |
| `x` | Cut to clipboard |
| `p` | Paste (asks confirmation) |

### Search

| Key | Action |
|---|---|
| `/` or `f` | Fuzzy file search by name |
| `ctrl+f` | Grep — search inside files |
| `↑` / `↓` | Navigate results |
| `enter` | Open selected result |
| `esc` | Exit search |

### Editor

| Key | Action |
|---|---|
| `ctrl+s` | Save |
| `esc` | Discard (asks confirmation if modified) |

### AI Agent

| Key | Action |
|---|---|
| `c` | Open AI chat (prompts for API key on first use) |
| `m` | Switch model (Opus / Sonnet / Haiku) |
| `enter` | Send message |
| `↑` / `↓` | Scroll conversation |
| `esc` | Close chat |

The agent can read, write, and create files in your project. Your API key is stored at `~/.config/beaver/apikey` and is never shared. Get a key at [console.anthropic.com](https://console.anthropic.com).

---

## Configuration

Beaver works without any config file. To customise, copy `configs/config.example.yaml` to `configs/config.yaml` next to the binary:

```yaml
ui:
  sidebar_width: 26
  show_key_hints: true
  scroll_speed: 3
  mouse_enabled: true
```

---

## Building from Source

```sh
git clone https://github.com/Beaver-family/beaver.git
cd beaver
go build -o beaver ./cmd/beaver
./beaver
```

Requires Go 1.21+.

---

## Contributing

Pull requests are welcome. For large changes, open an issue first to discuss what you'd like to change.

---

## Update

### Homebrew

```sh
brew update && brew upgrade beaver
```

### Go install

```sh
go install github.com/Beaver-family/beaver/cmd/beaver@latest
```

### Binary

Download the latest release from the [releases page](https://github.com/Beaver-family/beaver/releases/latest) and replace your existing binary.

---

## Uninstall

### Homebrew

```sh
brew uninstall beaver
brew untap Beaver-family/tap
```

### Go install

```sh
rm $(go env GOPATH)/bin/beaver
```

### Manual binary

Delete the `beaver` binary from wherever you placed it on your `$PATH`.

### Saved config and API key

Beaver stores your API key and model preference here — remove this directory to wipe everything:

```sh
rm -rf ~/.config/beaver
```

---

## License

MIT
