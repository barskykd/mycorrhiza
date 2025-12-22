package mdrenderer

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// Markdown is the configured goldmark instance
var Markdown goldmark.Markdown

func init() {
	Markdown = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,        // GitHub Flavored Markdown
			extension.Table,      // Tables
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(), // Respect line breaks
			html.WithXHTML(),     // XHTML-compliant output
		),
	)
}

// Render converts Markdown to HTML
func Render(source []byte) (string, error) {
	var buf bytes.Buffer
	if err := Markdown.Convert(source, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
