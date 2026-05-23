# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./...                          # build all packages
go run ./cmd/beaver                     # run the TUI directly
go build -o beaver ./cmd/beaver         # produce a named binary
go install github.com/Beaver-family/beaver/cmd/beaver@latest  # install
```

No linter or test suite is configured yet.

## Architecture

Beaver is a bubbletea TUI file manager. The entry point is `cmd/beaver/main.go`. All TUI logic lives under `internal/ui/`.

### Bubbletea model split

The bubbletea `Model` is split across three files:
- `model.go` — struct definition and all state fields
- `init.go` — `New()` constructor (starts at `os.Getwd()`) and `Init()` tick command
- `update.go` — entire `Update()` dispatch + all sub-handlers
- `view.go` — entire `View()` render + all sub-renderers

### State machine in update.go

`Update()` dispatches in strict priority order:
1. `opMode != opNone || editConfirm` → `updateOp()` (file op confirm/input prompts)
2. `editMode` → `updateEditor()`
3. `searchMode` → `updateSearch()`
4. `grepMode` → `updateGrep()`
5. Normal navigation keys

`opMode` is an `int` enum: `opNone / opConfirmDelete / opConfirmPaste / opInputRename / opInputNewFile / opInputNewDir`.

Grep has a two-phase state machine inside `grepMode`: `grepResults == nil` = input phase, `grepResults != nil` (even empty) = results phase.

### Layout

Two panels side by side: sidebar (file tree) | main (preview / editor / grep results).  
`panelHeight = m.height - 2` (status bar is 2 lines).  
`mainWidth = m.width - sidebarWidth - 2` (each border eats 1 col).

Status bar is always 2 lines at the bottom. When `opMode != opNone || editConfirm`, the status bar becomes a nano-style inverted prompt (see `renderOpStatusBar`).

### Sub-packages

| Package | Purpose |
|---|---|
| `internal/ui/filetree` | Lazy-loading tree (`Tree`/`Node`). `Flat []* Node` is the rendered view — only expanded nodes appear. `Refresh(dirPath)` reloads a subtree in-place. |
| `internal/ui/preview` | Async `tea.Cmd` that reads + syntax-highlights via chroma (Dracula/terminal256). Rejects binary files and files > 500 KB. |
| `internal/ui/editor` | Thin wrapper around `bubbles/textarea` with theme styling. |
| `internal/ui/fileops` | Each operation is a `tea.Cmd` returning `DoneMsg` or `ErrMsg`. Move tries `os.Rename` first (fast, same-fs), falls back to copy+delete. `DoneMsg.SrcDir` is populated for moves so both source and destination dirs are refreshed. |
| `internal/ui/search` | `FuzzyFiles` — synchronous filename search. `GrepCmd` — async `tea.Cmd`, caps at 500 results, skips hidden dirs and binary files. |
| `internal/ui/styles` | Single `theme.go` — all colours and lipgloss base styles. Sidebar width comes from config. |
| `internal/config` | Viper-based config loaded from `configs/config.yaml`. Defaults are set in code so the file is optional. Only `UIConfig` (sidebar width, hints, scroll speed, mouse) is active. |

### Clipboard

Two parallel representations must be kept in sync:
- `model.clipboard *filetree.Node` + `model.clipCut bool` — used for the paste operation
- `model.filetree.ClipboardPath string` + `model.filetree.ClipboardCut bool` — used by the filetree renderer to colour the item (red = cut, teal = copy)

Both must be set together on `y`/`x` and cleared together after a cut-paste.

### Release

GoReleaser v2 config is in `.goreleaser.yaml`. Builds for darwin (arm64/amd64), linux (arm64/amd64), windows (amd64). Triggered by pushing a `v*` tag — GitHub Actions workflow at `.github/workflows/release.yml` handles it. Homebrew cask is pushed to `Beaver-family/homebrew-tap` using `HOMEBREW_TAP_TOKEN` secret.
