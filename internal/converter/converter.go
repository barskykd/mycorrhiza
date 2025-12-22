package converter

import (
	"regexp"
	"strings"

	"github.com/bouncepaw/mycorrhiza/internal/hyphae"
)

// ConvertMycomarkupToMarkdown attempts to convert Mycomarkup to Markdown
// This is LOSSY - some features don't have Markdown equivalents
func MycomarkupToMarkdown(content string) (string, []string) {
	warnings := []string{}
	result := content

	// Headings: = Heading -> # Heading
	result = regexp.MustCompile(`(?m)^===\s+(.+)$`).ReplaceAllString(result, "### $1")
	result = regexp.MustCompile(`(?m)^==\s+(.+)$`).ReplaceAllString(result, "## $1")
	result = regexp.MustCompile(`(?m)^=\s+(.+)$`).ReplaceAllString(result, "# $1")

	// Links: [[link]] -> [link](link), [[link | text]] -> [text](link)
	result = regexp.MustCompile(`\[\[([^|\]]+)\|([^\]]+)\]\]`).ReplaceAllString(result, "[$2]($1)")
	result = regexp.MustCompile(`\[\[([^\]]+)\]\]`).ReplaceAllString(result, "[$1]($1)")

	// Bold: **text** -> **text** (same!)
	// Italic: //text// -> *text*
	result = regexp.MustCompile(`//([^/]+)//`).ReplaceAllString(result, "*$1*")

	// Monospace: `code` -> `code` (same!)

	// Lists: * item -> - item
	result = regexp.MustCompile(`(?m)^\*\s+`).ReplaceAllString(result, "- ")
	// Numbered: *. item -> 1. item
	result = regexp.MustCompile(`(?m)^\*\.\s+`).ReplaceAllString(result, "1. ")

	// Horizontal bar: ---- -> ---
	result = regexp.MustCompile(`(?m)^----+$`).ReplaceAllString(result, "---")

	// Code blocks: ``` stays the same

	// Features that don't convert well:
	if strings.Contains(content, "++") {
		warnings = append(warnings, "Highlighting (++) not supported in Markdown")
	}
	if strings.Contains(content, "^^") {
		warnings = append(warnings, "Superscript (^^) not supported in standard Markdown")
	}
	if strings.Contains(content, ",,") {
		warnings = append(warnings, "Subscript (,,) not supported in standard Markdown")
	}
	if strings.Contains(content, "__") {
		warnings = append(warnings, "Underline (__) not supported in Markdown")
	}
	if strings.Contains(content, "=>") {
		warnings = append(warnings, "Rocket links (=>) have no Markdown equivalent")
	}
	if strings.Contains(content, "<=") {
		warnings = append(warnings, "Transclusions (<=) have no Markdown equivalent")
	}
	if regexp.MustCompile(`(?m)^img\s*\{`).MatchString(content) {
		warnings = append(warnings, "Image blocks may need manual conversion to Markdown syntax")
	}
	if regexp.MustCompile(`(?m)^table\s*\{`).MatchString(content) {
		warnings = append(warnings, "Table blocks need manual conversion to Markdown tables")
	}

	return result, warnings
}

// ConvertMarkdownToMycomarkup attempts to convert Markdown to Mycomarkup
func MarkdownToMycomarkup(content string) (string, []string) {
	warnings := []string{}
	result := content

	// Headings: # Heading -> = Heading
	result = regexp.MustCompile(`(?m)^###\s+(.+)$`).ReplaceAllString(result, "=== $1")
	result = regexp.MustCompile(`(?m)^##\s+(.+)$`).ReplaceAllString(result, "== $1")
	result = regexp.MustCompile(`(?m)^#\s+(.+)$`).ReplaceAllString(result, "= $1")

	// Links: [text](url) -> [[url | text]]
	result = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(result, "[[$2 | $1]]")

	// Italic: *text* or _text_ -> //text//
	result = regexp.MustCompile(`\*([^*\n]+)\*`).ReplaceAllString(result, "//$1//")
	result = regexp.MustCompile(`_([^_\n]+)_`).ReplaceAllString(result, "//$1//")

	// Bold: **text** stays **text**

	// Lists: - item -> * item
	result = regexp.MustCompile(`(?m)^-\s+`).ReplaceAllString(result, "* ")

	// Numbered lists: convert to *.
	result = regexp.MustCompile(`(?m)^\d+\.\s+`).ReplaceAllString(result, "*. ")

	// Horizontal rule: --- -> ----
	result = regexp.MustCompile(`(?m)^---+$`).ReplaceAllString(result, "----")

	// Blockquote warning
	if strings.Contains(content, "\n>") || strings.HasPrefix(content, ">") {
		warnings = append(warnings, "Blockquotes (>) may need manual conversion")
	}

	// Code blocks: ``` stays the same

	if strings.Contains(content, "![") {
		warnings = append(warnings, "Image syntax ![...](...) may need conversion to img{} blocks")
	}

	return result, warnings
}

// ConvertFormat is the main entry point for format conversion
func ConvertFormat(content string, from, to hyphae.TextFormat) (string, []string, error) {
	if from == to {
		return content, []string{"No conversion needed - already in target format"}, nil
	}

	if from == hyphae.FormatMycomarkup && to == hyphae.FormatMarkdown {
		result, warnings := MycomarkupToMarkdown(content)
		return result, warnings, nil
	}

	if from == hyphae.FormatMarkdown && to == hyphae.FormatMycomarkup {
		result, warnings := MarkdownToMycomarkup(content)
		return result, warnings, nil
	}

	return content, []string{"Unknown conversion"}, nil
}
