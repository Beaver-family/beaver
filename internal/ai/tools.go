package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// BuildTools returns the tool definitions sent to Claude.
func BuildTools() []anthropic.ToolUnionParam {
	prop := func(desc string) map[string]any {
		return map[string]any{"type": "string", "description": desc}
	}
	mkTool := func(name string, props map[string]any, required []string) anthropic.ToolUnionParam {
		return anthropic.ToolUnionParam{OfTool: &anthropic.ToolParam{
			Name: name,
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: props,
				Required:   required,
			},
		}}
	}

	return []anthropic.ToolUnionParam{
		mkTool("read_file",
			map[string]any{"path": prop("Absolute path to the file to read")},
			[]string{"path"}),
		mkTool("write_file",
			map[string]any{
				"path":    prop("Absolute path to write"),
				"content": prop("Full new content of the file"),
			},
			[]string{"path", "content"}),
		mkTool("create_file",
			map[string]any{
				"path":    prop("Absolute path for the new file"),
				"content": prop("Initial content (may be empty)"),
			},
			[]string{"path"}),
		mkTool("list_directory",
			map[string]any{"path": prop("Absolute path to the directory")},
			[]string{"path"}),
		mkTool("search_files",
			map[string]any{
				"query":     prop("Filename query (case-insensitive substring)"),
				"root_path": prop("Directory to search under"),
			},
			[]string{"query", "root_path"}),
		mkTool("grep_content",
			map[string]any{
				"query":     prop("Text to search for inside files"),
				"root_path": prop("Directory to search under"),
			},
			[]string{"query", "root_path"}),
	}
}

// ExecuteTool runs a single tool call and returns the result string and whether it errored.
func ExecuteTool(name string, inputJSON json.RawMessage, rootPath string) (string, bool) {
	var args map[string]string
	if err := json.Unmarshal(inputJSON, &args); err != nil {
		return "invalid tool input: " + err.Error(), true
	}

	switch name {
	case "read_file":
		return toolReadFile(args["path"])
	case "write_file":
		return toolWriteFile(args["path"], args["content"])
	case "create_file":
		return toolCreateFile(args["path"], args["content"])
	case "list_directory":
		return toolListDir(args["path"])
	case "search_files":
		root := args["root_path"]
		if root == "" {
			root = rootPath
		}
		return toolSearchFiles(args["query"], root)
	case "grep_content":
		root := args["root_path"]
		if root == "" {
			root = rootPath
		}
		return toolGrepContent(args["query"], root)
	default:
		return "unknown tool: " + name, true
	}
}

func toolReadFile(path string) (string, bool) {
	if path == "" {
		return "path required", true
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err.Error(), true
	}
	content := string(data)
	const maxChars = 12000
	if len(content) > maxChars {
		content = content[:maxChars] + fmt.Sprintf("\n\n[truncated — file is %d bytes total]", len(data))
	}
	return content, false
}

func toolWriteFile(path, content string) (string, bool) {
	if path == "" {
		return "path required", true
	}
	perm := os.FileMode(0644)
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode()
	}
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		return err.Error(), true
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(content), filepath.Base(path)), false
}

func toolCreateFile(path, content string) (string, bool) {
	if path == "" {
		return "path required", true
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err.Error(), true
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err.Error(), true
	}
	return "created " + filepath.Base(path), false
}

func toolListDir(path string) (string, bool) {
	if path == "" {
		return "path required", true
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err.Error(), true
	}
	var sb strings.Builder
	for _, e := range entries {
		if e.Name()[0] == '.' {
			continue
		}
		if e.IsDir() {
			fmt.Fprintf(&sb, "  [dir]  %s\n", e.Name())
		} else {
			info, _ := e.Info()
			if info != nil {
				fmt.Fprintf(&sb, "  [file] %s  (%d B)\n", e.Name(), info.Size())
			} else {
				fmt.Fprintf(&sb, "  [file] %s\n", e.Name())
			}
		}
	}
	if sb.Len() == 0 {
		return "empty directory", false
	}
	return sb.String(), false
}

func toolSearchFiles(query, root string) (string, bool) {
	if query == "" {
		return "query required", true
	}
	lower := strings.ToLower(query)
	var matches []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.Name()[0] == '.' {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.Contains(strings.ToLower(d.Name()), lower) {
			matches = append(matches, path)
		}
		if len(matches) >= 100 {
			return filepath.SkipAll
		}
		return nil
	})
	if len(matches) == 0 {
		return "no files matching \"" + query + "\"", false
	}
	return strings.Join(matches, "\n"), false
}

func toolGrepContent(query, root string) (string, bool) {
	if query == "" {
		return "query required", true
	}
	lower := strings.ToLower(query)
	type hit struct {
		path    string
		lineNum int
		line    string
	}
	var results []hit
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name()[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name()[0] == '.' {
			return nil
		}
		info, _ := d.Info()
		if info != nil && info.Size() > 512*1024 {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.Contains(strings.ToLower(line), lower) {
				results = append(results, hit{path, i + 1, strings.TrimSpace(line)})
				if len(results) >= 200 {
					return filepath.SkipAll
				}
			}
		}
		return nil
	})
	if len(results) == 0 {
		return "no matches for \"" + query + "\"", false
	}
	var sb strings.Builder
	for _, r := range results {
		fmt.Fprintf(&sb, "  %s:%d  %s\n", filepath.Base(r.path), r.lineNum, r.line)
	}
	return sb.String(), false
}
