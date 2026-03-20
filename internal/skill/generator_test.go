package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/smilemilks2021/easy-web/internal/browser"
)

func TestDomainToSkillName(t *testing.T) {
	cases := []struct {
		url  string
		want string
	}{
		{"https://s-power.sheincorp.cn/task-list", "s-power"},
		{"https://grafana.internal.com/d/dashboard", "grafana"},
		{"https://example.com", "example"},
		{"https://sub.domain.example.org/path", "sub"},
		{"not-a-url", "unknown"},
	}
	for _, c := range cases {
		got := DomainToSkillName(c.url)
		if got != c.want {
			t.Errorf("DomainToSkillName(%q) = %q, want %q", c.url, got, c.want)
		}
	}
}

func TestExtractRecordedURLs(t *testing.T) {
	content := `## 接口列表

### GET /api/tasks
` + "```bash" + `
easy-web request -u "https://example.com/api/tasks"
` + "```" + `

### POST /api/tasks
` + "```bash" + `
easy-web request -u "https://example.com/api/tasks" \
  -X POST \
  -d '{"name":"test"}'
` + "```"

	urls := extractRecordedURLs(content)
	if !urls["https://example.com/api/tasks"] {
		t.Error("expected https://example.com/api/tasks to be extracted")
	}
	if len(urls) != 1 {
		t.Errorf("expected 1 unique URL, got %d: %v", len(urls), urls)
	}
}

func TestGenerate_WriteNew(t *testing.T) {
	dir := t.TempDir()
	// Override SkillDir by patching — use a test path via env trick
	// Instead, call writeNew directly
	reqs := []*browser.CapturedRequest{
		{URL: "https://example.com/api/tasks", Method: "GET"},
		{URL: "https://example.com/api/tasks", Method: "POST", Body: `{"name":"test"}`},
	}

	path := filepath.Join(dir, "example", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}

	if err := writeNew(path, reqs, "https://example.com", "example"); err != nil {
		t.Fatalf("writeNew error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)

	// Verify frontmatter
	if !strings.Contains(content, "name: example") {
		t.Error("missing name in frontmatter")
	}
	if !strings.Contains(content, "allowed-tools: Bash(easy-web:*)") {
		t.Error("missing allowed-tools")
	}
	if !strings.Contains(content, "Use when automating example") {
		t.Error("missing Use when... in description")
	}

	// Verify auth section
	if !strings.Contains(content, "easy-web -u https://example.com") {
		t.Error("missing auth command")
	}

	// Verify APIs
	if !strings.Contains(content, "### GET /api/tasks") {
		t.Error("missing GET API block")
	}
	if !strings.Contains(content, "### POST /api/tasks") {
		t.Error("missing POST API block")
	}
	if !strings.Contains(content, `'{"name":"test"}'`) {
		t.Error("missing POST body (expected single-quoted JSON)")
	}
}

func TestGenerate_SmartMerge(t *testing.T) {
	dir := t.TempDir()

	existing := `---
name: example
description: |
  Use when automating example (example.com)
allowed-tools: Bash(easy-web:*)
---

# example

## 认证
` + "```bash" + `
easy-web -u https://example.com
` + "```" + `

## 接口列表

### GET /api/tasks
` + "```bash" + `
easy-web request -u "https://example.com/api/tasks"
` + "```" + `
`

	path := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	// One existing + one new request
	reqs := []*browser.CapturedRequest{
		{URL: "https://example.com/api/tasks", Method: "GET"},      // already exists
		{URL: "https://example.com/api/users", Method: "GET"},      // new
	}

	if err := smartMerge(path, reqs, "https://example.com", "example"); err != nil {
		t.Fatalf("smartMerge error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	// Existing content preserved
	if !strings.Contains(content, "### GET /api/tasks") {
		t.Error("existing API should be preserved")
	}
	// New API added
	if !strings.Contains(content, "### GET /api/users") {
		t.Error("new API should be added")
	}
	// Count occurrences of /api/tasks — should only appear once
	count := strings.Count(content, "/api/tasks")
	if count != 2 { // once in heading, once in command
		t.Errorf("expected /api/tasks to appear exactly twice (heading + command), got %d", count)
	}
}
