package converter

import (
	"strings"
	"testing"

	"github.com/bouncepaw/mycorrhiza/internal/hyphae"
)

func TestMycomarkupToMarkdown_Headings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "level 1 heading",
			input: "= Heading 1",
			want:  "# Heading 1\n",
		},
		{
			name:  "level 2 heading",
			input: "== Heading 2",
			want:  "## Heading 2\n",
		},
		{
			name:  "level 3 heading",
			input: "=== Heading 3",
			want:  "### Heading 3\n",
		},
		{
			name:  "multiple headings",
			input: "= Heading 1\n\n== Heading 2\n\n=== Heading 3",
			want:  "# Heading 1\n\n## Heading 2\n\n### Heading 3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MycomarkupToMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("MycomarkupToMarkdown() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestMycomarkupToMarkdown_Formatting(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "italic text",
			input: "//italic text//",
			want:  "*italic text*\n",
		},
		{
			name:  "bold text",
			input: "**bold text**",
			want:  "**bold text**\n",
		},
		{
			name:  "monospace text",
			input: "`code text`",
			want:  "`code text`\n",
		},
		{
			name:  "bold and italic",
			input: "**//bold italic//**",
			want:  "***bold italic***\n",
		},
		{
			name:  "plain paragraph",
			input: "This is plain text.",
			want:  "This is plain text.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MycomarkupToMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("MycomarkupToMarkdown() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestMycomarkupToMarkdown_Links(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple link",
			input: "[[Page]]",
			want:  "[Page](Page)\n",
		},
		{
			name:  "link with custom text",
			input: "[[Page | Custom Text]]",
			// Note: mycomarkup normalizes hypha names to lowercase, so we get "page" not "Page"
			want:  "[Custom Text](page)\n",
		},
		{
			name:  "link in paragraph",
			input: "Check out [[Page]] for more info.",
			want:  "Check out [Page](Page) for more info.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MycomarkupToMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("MycomarkupToMarkdown() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestMycomarkupToMarkdown_Lists(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "unordered list",
			input: "* Item 1\n* Item 2\n* Item 3",
			want:  "- Item 1\n- Item 2\n- Item 3\n",
		},
		{
			name:  "ordered list",
			input: "*. First\n*. Second\n*. Third",
			want:  "1. First\n1. Second\n1. Third\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MycomarkupToMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("MycomarkupToMarkdown() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestMycomarkupToMarkdown_CodeBlocks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "code block without language",
			input: "```\ncode here\n```",
			want:  "```\ncode here\n```\n",
		},
		{
			name:  "code block with language",
			input: "```go\nfunc main() {}\n```",
			want:  "```go\nfunc main() {}\n```\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MycomarkupToMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("MycomarkupToMarkdown() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestMycomarkupToMarkdown_ThematicBreak(t *testing.T) {
	input := "----"
	want := "---\n"

	got, _ := MycomarkupToMarkdown(input)
	if got != want {
		t.Errorf("MycomarkupToMarkdown() = %q, want %q", got, want)
	}
}

func TestMycomarkupToMarkdown_HTMLFallbacks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "superscript",
			input: "E = mc^^2^^",
			want:  "E = mc<sup>2</sup>\n",
		},
		{
			name:  "subscript",
			input: "H,,2,,O",
			want:  "H<sub>2</sub>O\n",
		},
		{
			name:  "underline",
			input: "__underlined text__",
			want:  "<u>underlined text</u>\n",
		},
		{
			name:  "highlight",
			input: "++highlighted text++",
			want:  "<mark>highlighted text</mark>\n",
		},
		{
			name:  "strikethrough",
			input: "~~strikethrough~~",
			want:  "~~strikethrough~~\n",
		},
		{
			name:  "combined formatting",
			input: "**bold** and //italic// and ~~strike~~",
			want:  "**bold** and *italic* and ~~strike~~\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MycomarkupToMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("MycomarkupToMarkdown() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestMarkdownToMycomarkup_Headings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "level 1 heading",
			input: "# Heading 1",
			want:  "= Heading 1\n",
		},
		{
			name:  "level 2 heading",
			input: "## Heading 2",
			want:  "== Heading 2\n",
		},
		{
			name:  "level 3 heading",
			input: "### Heading 3",
			want:  "=== Heading 3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MarkdownToMycomarkup(tt.input)
			if got != tt.want {
				t.Errorf("MarkdownToMycomarkup() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestMarkdownToMycomarkup_Formatting(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "italic with asterisk",
			input: "*italic*",
			want:  "//italic//\n",
		},
		{
			name:  "italic with underscore",
			input: "_italic_",
			want:  "//italic//\n",
		},
		{
			name:  "bold",
			input: "**bold**",
			want:  "**bold**\n",
		},
		{
			name:  "code",
			input: "`code`",
			want:  "`code`\n",
		},
		{
			name:  "bold and italic",
			input: "***bold italic***",
			want:  "**//bold italic//**\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MarkdownToMycomarkup(tt.input)
			if got != tt.want {
				t.Errorf("MarkdownToMycomarkup() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestMarkdownToMycomarkup_Links(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "link with same text and URL",
			input: "[Page](Page)",
			want:  "[[Page]]\n",
		},
		{
			name:  "link with different text",
			input: "[Custom Text](Page)",
			want:  "[[Page | Custom Text]]\n",
		},
		{
			name:  "link with URL",
			input: "[Google](https://google.com)",
			want:  "[[https://google.com | Google]]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MarkdownToMycomarkup(tt.input)
			if got != tt.want {
				t.Errorf("MarkdownToMycomarkup() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestMarkdownToMycomarkup_Lists(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "unordered list",
			input: "- Item 1\n- Item 2\n- Item 3",
			want:  "* Item 1\n* Item 2\n* Item 3\n",
		},
		{
			name:  "ordered list",
			input: "1. First\n2. Second\n3. Third",
			want:  "*. First\n*. Second\n*. Third\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MarkdownToMycomarkup(tt.input)
			if got != tt.want {
				t.Errorf("MarkdownToMycomarkup() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestMarkdownToMycomarkup_CodeBlocks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "code block without language",
			input: "```\ncode here\n```",
			want:  "```\ncode here\n```\n",
		},
		{
			name:  "code block with language",
			input: "```go\nfunc main() {}\n```",
			want:  "```go\nfunc main() {}\n```\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MarkdownToMycomarkup(tt.input)
			if got != tt.want {
				t.Errorf("MarkdownToMycomarkup() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestMarkdownToMycomarkup_ThematicBreak(t *testing.T) {
	input := "---"
	want := "----\n"

	got, _ := MarkdownToMycomarkup(input)
	if got != want {
		t.Errorf("MarkdownToMycomarkup() = %q, want %q", got, want)
	}
}

func TestConvertFormat(t *testing.T) {
	tests := []struct {
		name    string
		content string
		from    hyphae.TextFormat
		to      hyphae.TextFormat
		want    string
	}{
		{
			name:    "same format - no conversion",
			content: "test",
			from:    hyphae.FormatMarkdown,
			to:      hyphae.FormatMarkdown,
			want:    "test",
		},
		{
			name:    "myco to markdown",
			content: "= Heading",
			from:    hyphae.FormatMycomarkup,
			to:      hyphae.FormatMarkdown,
			want:    "# Heading\n",
		},
		{
			name:    "markdown to myco",
			content: "# Heading",
			from:    hyphae.FormatMarkdown,
			to:      hyphae.FormatMycomarkup,
			want:    "= Heading\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := ConvertFormat(tt.content, tt.from, tt.to)
			if err != nil {
				t.Errorf("ConvertFormat() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ConvertFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		// After round trip myco->md->myco, we expect this output
		wantMycoOutput string
	}{
		{
			name:           "heading",
			input:          "= Heading",
			wantMycoOutput: "= Heading\n",
		},
		{
			name:           "bold text",
			input:          "**bold**",
			wantMycoOutput: "**bold**\n",
		},
		{
			name:           "italic text",
			input:          "//italic//",
			wantMycoOutput: "//italic//\n",
		},
		{
			name:           "code",
			input:          "`code`",
			wantMycoOutput: "`code`\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert mycomarkup -> markdown
			md, _ := MycomarkupToMarkdown(tt.input)

			// Convert markdown -> mycomarkup
			myco, _ := MarkdownToMycomarkup(md)

			if myco != tt.wantMycoOutput {
				t.Errorf("Round trip failed:\nInput:  %q\nMD:     %q\nOutput: %q\nWant:   %q",
					tt.input, md, myco, tt.wantMycoOutput)
			}
		})
	}
}

func TestMycomarkupToMarkdown_Images(t *testing.T) {
	// Note: These tests verify the conversion produces valid output
	// Exact output format may vary based on image block structure
	tests := []struct {
		name      string
		input     string
		wantMatch string // substring that should be in output
	}{
		{
			name:      "simple image",
			input:     "img { test.png }",
			wantMatch: "![",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MycomarkupToMarkdown(tt.input)
			if !strings.Contains(got, tt.wantMatch) {
				t.Errorf("MycomarkupToMarkdown() = %q, want to contain %q", got, tt.wantMatch)
			}
		})
	}
}

func TestMycomarkupToMarkdown_Tables(t *testing.T) {
	// Tables should convert to HTML
	input := `table {
th { Header 1 }
th { Header 2 }
Cell 1
Cell 2
}`

	got, _ := MycomarkupToMarkdown(input)

	// Should contain HTML table elements
	if !strings.Contains(got, "<table>") {
		t.Errorf("Expected HTML table output, got: %q", got)
	}
	if !strings.Contains(got, "</table>") {
		t.Errorf("Expected closing table tag, got: %q", got)
	}
}

func TestMarkdownToMycomarkup_Strikethrough(t *testing.T) {
	input := "~~strikethrough text~~"
	want := "~~strikethrough text~~\n"

	got, _ := MarkdownToMycomarkup(input)
	if got != want {
		t.Errorf("MarkdownToMycomarkup() = %q, want %q", got, want)
	}
}

func TestMarkdownToMycomarkup_Images(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "image with alt text",
			input: "![Alt text](image.png)",
			want:  "img { image.png | Alt text }\n",
		},
		{
			name:  "image without alt",
			input: "![](image.png)",
			want:  "img { image.png }\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MarkdownToMycomarkup(tt.input)
			if got != tt.want {
				t.Errorf("MarkdownToMycomarkup() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMarkdownToMycomarkup_Blockquotes(t *testing.T) {
	input := "> This is a quote"
	got, _ := MarkdownToMycomarkup(input)

	// Should start with > for mycomarkup quote
	if !strings.HasPrefix(got, "> ") {
		t.Errorf("Expected blockquote to start with '> ', got: %q", got)
	}
}

func TestMarkdownToMycomarkup_Tables(t *testing.T) {
	input := `| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |`

	got, _ := MarkdownToMycomarkup(input)

	// Should convert to mycomarkup table syntax
	if !strings.Contains(got, "table {") {
		t.Errorf("Expected mycomarkup table syntax, got: %q", got)
	}
}

func TestRoundTrip_HTMLFeatures(t *testing.T) {
	tests := []struct {
		name  string
		input string
		// After round trip, should preserve meaning even if syntax changes
		wantContains string
	}{
		{
			name:         "superscript",
			input:        "x^^2^^",
			wantContains: "2", // Should preserve the content
		},
		{
			name:         "subscript",
			input:        "H,,2,,O",
			wantContains: "2", // Should preserve the content
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert myco -> md
			md, _ := MycomarkupToMarkdown(tt.input)

			// Convert md -> myco
			myco, _ := MarkdownToMycomarkup(md)

			if !strings.Contains(myco, tt.wantContains) {
				t.Errorf("Round trip lost content. Input: %q, MD: %q, Output: %q, Want to contain: %q",
					tt.input, md, myco, tt.wantContains)
			}
		})
	}
}
