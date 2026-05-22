package preview

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	tea "github.com/charmbracelet/bubbletea"
)

const maxBytes = 512 * 1024 // 500 KB

var unsupportedExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".bmp": true, ".ico": true, ".tiff": true, ".svg": true,
	".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".webm": true,
	".mp3": true, ".flac": true, ".wav": true, ".aac": true, ".ogg": true,
	".zip": true, ".tar": true, ".gz": true, ".bz2": true, ".xz": true,
	".rar": true, ".7z": true,
	".pdf": true, ".docx": true, ".xlsx": true, ".pptx": true,
	".exe": true, ".dll": true, ".so": true, ".dylib": true, ".bin": true,
	".db": true, ".sqlite": true, ".sqlite3": true,
}

// LoadedMsg is sent to the Update loop when a file loads successfully.
type LoadedMsg struct {
	Path  string
	Lines []string
}

// ErrMsg is sent when a file cannot or should not be previewed.
type ErrMsg struct {
	Path string
	Text string
}

// Load returns a tea.Cmd that reads and syntax-highlights the file at path.
func Load(path string) tea.Cmd {
	return func() tea.Msg {
		ext := strings.ToLower(filepath.Ext(path))

		if unsupportedExts[ext] {
			return ErrMsg{Path: path, Text: "not supported"}
		}

		info, err := os.Stat(path)
		if err != nil {
			return ErrMsg{Path: path, Text: "cannot read file"}
		}
		if info.Size() > maxBytes {
			return ErrMsg{Path: path, Text: "file too large (> 500 KB)"}
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			return ErrMsg{Path: path, Text: "cannot read file"}
		}

		// sniff: reject non-UTF-8 or binary (null bytes)
		if !utf8.Valid(raw) || bytes.ContainsRune(raw, 0) {
			return ErrMsg{Path: path, Text: "not supported (binary)"}
		}

		// expand tabs before highlighting so width calculations are exact
		src := strings.ReplaceAll(string(raw), "\t", "    ")
		highlighted := highlight(path, src)
		lines := strings.Split(highlighted, "\n")
		return LoadedMsg{Path: path, Lines: lines}
	}
}

func highlight(path, src string) string {
	lexer := lexers.Match(path)
	if lexer == nil {
		lexer = lexers.Analyse(src)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("dracula")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	it, err := lexer.Tokenise(nil, src)
	if err != nil {
		return src
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, it); err != nil {
		return src
	}
	return buf.String()
}
