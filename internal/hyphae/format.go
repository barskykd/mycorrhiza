package hyphae

import (
	"path/filepath"
)

// TextFormat represents the markup format of a hypha's text content
type TextFormat int

const (
	FormatMycomarkup TextFormat = iota
	FormatMarkdown
)

// DetectTextFormat returns the format based on file extension
func DetectTextFormat(filePath string) TextFormat {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".md":
		return FormatMarkdown
	case ".myco":
		return FormatMycomarkup
	default:
		return FormatMycomarkup // default for backward compatibility
	}
}

// FormatExtension returns the file extension for a format
func FormatExtension(format TextFormat) string {
	switch format {
	case FormatMarkdown:
		return ".md"
	case FormatMycomarkup:
		return ".myco"
	default:
		return ".myco"
	}
}

// FormatName returns human-readable name
func FormatName(format TextFormat) string {
	switch format {
	case FormatMarkdown:
		return "Markdown"
	case FormatMycomarkup:
		return "Mycomarkup"
	default:
		return "Mycomarkup"
	}
}
