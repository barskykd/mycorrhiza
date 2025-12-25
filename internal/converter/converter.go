package converter

import (
	"fmt"
	"strings"

	"github.com/bouncepaw/mycorrhiza/internal/hyphae"
	"github.com/bouncepaw/mycorrhiza/mycoopts"

	"git.sr.ht/~bouncepaw/mycomarkup/v5"
	"git.sr.ht/~bouncepaw/mycomarkup/v5/blocks"
	"git.sr.ht/~bouncepaw/mycomarkup/v5/mycocontext"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	gast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// MycomarkupToMarkdown converts Mycomarkup to Markdown using AST parsing
// This is LOSSY - some features don't have Markdown equivalents
func MycomarkupToMarkdown(content string) (string, []string) {
	warnings := []string{}
	var output strings.Builder

	// Parse mycomarkup into AST
	ctx, _ := mycocontext.ContextFromStringInput(content, mycoopts.MarkupOptions(""))
	tree := mycomarkup.BlockTree(ctx)

	// Convert each block
	for i, block := range tree {
		if i > 0 {
			output.WriteString("\n")
		}
		convertMycoBlockToMarkdown(block, &output, &warnings, 0)
	}

	return output.String(), warnings
}

func convertMycoBlockToMarkdown(block blocks.Block, output *strings.Builder, warnings *[]string, depth int) {
	switch b := block.(type) {
	case blocks.Heading:
		// Mycomarkup: = Heading -> Markdown: # Heading
		output.WriteString(strings.Repeat("#", int(b.Level())) + " ")
		contents := b.Contents()
		convertFormattedToMarkdown(&contents, output, warnings)
		output.WriteString("\n")

	case blocks.Paragraph:
		convertFormattedToMarkdown(&b.Formatted, output, warnings)
		output.WriteString("\n")

	case blocks.CodeBlock:
		// Code blocks: ``` stays the same
		output.WriteString("```")
		lang := b.Language()
		// Skip "plain" as it's mycomarkup's default for unspecified language
		if lang != "" && lang != "plain" {
			output.WriteString(lang)
		}
		output.WriteString("\n")
		output.WriteString(b.Contents())
		if !strings.HasSuffix(b.Contents(), "\n") {
			output.WriteString("\n")
		}
		output.WriteString("```\n")

	case blocks.List:
		for _, item := range b.Items {
			convertListItemToMarkdown(item, output, warnings, depth, b.Marker)
		}

	case blocks.ThematicBreak:
		output.WriteString("---\n")

	case blocks.Quote:
		// Quotes: > in Markdown
		for _, subBlock := range b.Contents() {
			output.WriteString("> ")
			// Convert the sub-block inline
			var quoteOutput strings.Builder
			convertMycoBlockToMarkdown(subBlock, &quoteOutput, warnings, depth)
			// Remove trailing newline and add it back outside the quote marker
			quoted := strings.TrimSuffix(quoteOutput.String(), "\n")
			output.WriteString(quoted)
			output.WriteString("\n")
		}

	case blocks.Table:
		convertTableToHTML(b, output, warnings)

	case blocks.Img:
		convertImgToMarkdown(b, output, warnings)

	case blocks.LaunchPad:
		convertLaunchPadToHTML(b, output, warnings)

	case blocks.Transclusion:
		convertTransclusionToHTML(b, output, warnings)

	default:
		// Unknown block type, try to preserve as-is
		output.WriteString(fmt.Sprintf("<!-- Unknown block type: %T -->\n", block))
	}
}

func convertListItemToMarkdown(item blocks.ListItem, output *strings.Builder, warnings *[]string, depth int, marker blocks.ListMarker) {
	indent := strings.Repeat("  ", depth)

	// Convert marker
	switch marker {
	case blocks.MarkerUnordered:
		output.WriteString(indent + "- ")
	case blocks.MarkerOrdered:
		output.WriteString(indent + "1. ")
	}

	// Convert item contents - ListItem.Contents is []Block, not Formatted
	for _, block := range item.Contents {
		convertMycoBlockToMarkdown(block, output, warnings, depth)
	}
}

func convertFormattedToMarkdown(formatted *blocks.Formatted, output *strings.Builder, warnings *[]string) {
	// Track active styles
	styleState := blocks.CleanStyleState()

	for _, line := range formatted.Lines {
		for _, span := range line {
			switch s := span.(type) {
			case blocks.SpanTableEntry:
				// Toggle style
				kind := s.Kind()
				styleState[kind] = !styleState[kind]

			case blocks.InlineText:
				convertInlineTextToMarkdown(s, styleState, output, warnings)

			case blocks.InlineLink:
				convertInlineLinkToMarkdown(s, output, warnings)
			}
		}
	}
}

func convertInlineTextToMarkdown(text blocks.InlineText, styleState map[blocks.SpanKind]bool, output *strings.Builder, warnings *[]string) {
	// Use HTML tags for styles not supported in standard Markdown
	if styleState[blocks.SpanSuper] {
		output.WriteString("<sup>")
	}
	if styleState[blocks.SpanSub] {
		output.WriteString("<sub>")
	}
	if styleState[blocks.SpanUnderline] {
		output.WriteString("<u>")
	}
	if styleState[blocks.SpanMark] {
		output.WriteString("<mark>")
	}

	// Apply supported styles in order: bold, italic, strikethrough, monospace
	if styleState[blocks.SpanBold] {
		output.WriteString("**")
	}
	if styleState[blocks.SpanItalic] {
		output.WriteString("*")
	}
	if styleState[blocks.SpanStrike] {
		output.WriteString("~~")
	}
	if styleState[blocks.SpanMono] {
		output.WriteString("`")
	}

	output.WriteString(text.Contents)

	// Close styles in reverse order
	if styleState[blocks.SpanMono] {
		output.WriteString("`")
	}
	if styleState[blocks.SpanStrike] {
		output.WriteString("~~")
	}
	if styleState[blocks.SpanItalic] {
		output.WriteString("*")
	}
	if styleState[blocks.SpanBold] {
		output.WriteString("**")
	}

	// Close HTML tags in reverse order
	if styleState[blocks.SpanMark] {
		output.WriteString("</mark>")
	}
	if styleState[blocks.SpanUnderline] {
		output.WriteString("</u>")
	}
	if styleState[blocks.SpanSub] {
		output.WriteString("</sub>")
	}
	if styleState[blocks.SpanSuper] {
		output.WriteString("</sup>")
	}
}

func convertInlineLinkToMarkdown(link blocks.InlineLink, output *strings.Builder, warnings *[]string) {
	// [[link]] or [[link | text]]
	// InlineLink embeds links.Link which has methods LinkHref() and DisplayedText()
	// We need a context to get the href, but we'll use a minimal one
	ctx, _ := mycocontext.ContextFromStringInput("", mycoopts.MarkupOptions(""))

	href := link.LinkHref(ctx)
	displayText := link.DisplayedText()

	// Strip /hypha/ prefix if present to get just the page name
	href = strings.TrimPrefix(href, "/hypha/")

	// If displayText matches href (case-insensitive), use displayText to preserve case
	if strings.EqualFold(displayText, href) {
		href = displayText
	}

	// [text](href)
	output.WriteString("[")
	output.WriteString(displayText)
	output.WriteString("](")
	output.WriteString(href)
	output.WriteString(")")
}

func convertTableToHTML(table blocks.Table, output *strings.Builder, warnings *[]string) {
	output.WriteString("<table>\n")

	if caption := table.Caption(); caption != "" {
		output.WriteString("<caption>")
		output.WriteString(caption)
		output.WriteString("</caption>\n")
	}

	rows := table.Rows()
	for i, row := range rows {
		// First row might be header
		isHeader := i == 0 && row.LooksLikeThead()
		if isHeader {
			output.WriteString("<thead>\n")
		}

		output.WriteString("<tr>")
		for _, cell := range row.Cells() {
			tag := "td"
			if cell.IsHeaderCell() || isHeader {
				tag = "th"
			}

			output.WriteString("<" + tag)
			if colspan := cell.Colspan(); colspan > 1 {
				output.WriteString(fmt.Sprintf(" colspan=\"%d\"", colspan))
			}
			output.WriteString(">")

			// Convert cell contents
			for _, block := range cell.Contents() {
				var cellOutput strings.Builder
				convertMycoBlockToMarkdown(block, &cellOutput, warnings, 0)
				// Remove trailing newline for inline display
				content := strings.TrimSuffix(cellOutput.String(), "\n")
				output.WriteString(content)
			}

			output.WriteString("</" + tag + ">")
		}
		output.WriteString("</tr>\n")

		if isHeader {
			output.WriteString("</thead>\n<tbody>\n")
		}
	}

	if len(rows) > 0 && rows[0].LooksLikeThead() {
		output.WriteString("</tbody>\n")
	}

	output.WriteString("</table>\n")
}

func convertImgToMarkdown(img blocks.Img, output *strings.Builder, warnings *[]string) {
	ctx, _ := mycocontext.ContextFromStringInput("", mycoopts.MarkupOptions(""))

	// If single image, use markdown syntax
	if img.HasOneImage() && len(img.Entries) == 1 {
		entry := img.Entries[0]
		src := entry.Target.ImgSrc(ctx)

		// Get description if available
		alt := ""
		if desc := entry.Description(); len(desc) > 0 {
			var descOutput strings.Builder
			for _, block := range desc {
				convertMycoBlockToMarkdown(block, &descOutput, warnings, 0)
			}
			alt = strings.TrimSpace(descOutput.String())
		}

		output.WriteString("![")
		output.WriteString(alt)
		output.WriteString("](")
		output.WriteString(src)
		output.WriteString(")\n")
	} else {
		// Multiple images or complex layout - use HTML
		output.WriteString("<div class=\"img-gallery\">\n")
		for _, entry := range img.Entries {
			src := entry.Target.ImgSrc(ctx)
			output.WriteString("<img src=\"")
			output.WriteString(src)
			output.WriteString("\"")

			if w := entry.Width(); w != "" {
				output.WriteString(" width=\"" + w + "\"")
			}
			if h := entry.Height(); h != "" {
				output.WriteString(" height=\"" + h + "\"")
			}

			if desc := entry.Description(); len(desc) > 0 {
				var descOutput strings.Builder
				for _, block := range desc {
					convertMycoBlockToMarkdown(block, &descOutput, warnings, 0)
				}
				alt := strings.TrimSpace(descOutput.String())
				output.WriteString(" alt=\"" + alt + "\"")
			}

			output.WriteString(">\n")
		}
		output.WriteString("</div>\n")
	}
}

func convertLaunchPadToHTML(lp blocks.LaunchPad, output *strings.Builder, warnings *[]string) {
	ctx, _ := mycocontext.ContextFromStringInput("", mycoopts.MarkupOptions(""))

	output.WriteString("<div class=\"launchpad\">\n")
	for _, rocket := range lp.Rockets {
		if rocket.IsEmpty {
			continue
		}

		href := rocket.LinkHref(ctx)
		text := rocket.DisplayedText()

		output.WriteString("<a href=\"")
		output.WriteString(href)
		output.WriteString("\">")
		output.WriteString(text)
		output.WriteString("</a><br>\n")
	}
	output.WriteString("</div>\n")
}

func convertTransclusionToHTML(t blocks.Transclusion, output *strings.Builder, warnings *[]string) {
	// Transclusions can't be converted - preserve as HTML comment
	output.WriteString("<!-- Transclusion: ")
	// We can't easily get the original syntax, so just note it
	output.WriteString("transclusion not supported in Markdown")
	output.WriteString(" -->\n")
}

// MarkdownToMycomarkup converts Markdown to Mycomarkup using AST parsing
func MarkdownToMycomarkup(content string) (string, []string) {
	warnings := []string{}
	var output strings.Builder

	// Create goldmark parser with extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
		),
	)

	// Parse markdown
	source := []byte(content)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	// Walk the AST
	firstBlock := true
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		// Only process on entering (not leaving) nodes
		if !entering {
			return ast.WalkContinue, nil
		}

		// Skip the document root
		if node.Kind() == ast.KindDocument {
			return ast.WalkContinue, nil
		}

		// Add blank line between blocks (except first)
		if node.Type() == ast.TypeBlock && !firstBlock && node.Parent().Kind() == ast.KindDocument {
			output.WriteString("\n")
		}
		if node.Type() == ast.TypeBlock && node.Parent().Kind() == ast.KindDocument {
			firstBlock = false
		}

		convertMarkdownNodeToMycomarkup(node, source, &output, &warnings)

		return ast.WalkContinue, nil
	})

	return output.String(), warnings
}

func convertMarkdownNodeToMycomarkup(node ast.Node, source []byte, output *strings.Builder, warnings *[]string) {
	switch n := node.(type) {
	case *ast.Heading:
		// # Heading -> = Heading
		output.WriteString(strings.Repeat("=", n.Level) + " ")
		convertMarkdownInlineChildren(n, source, output, warnings)
		output.WriteString("\n")

	case *ast.Paragraph:
		// Only process top-level paragraphs
		// Paragraphs in list items are handled by the ListItem case
		if n.Parent().Kind() == ast.KindDocument {
			convertMarkdownInlineChildren(n, source, output, warnings)
			output.WriteString("\n")
		}

	case *ast.FencedCodeBlock:
		output.WriteString("```")
		if n.Language(source) != nil {
			output.Write(n.Language(source))
		}
		output.WriteString("\n")
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			output.Write(line.Value(source))
		}
		output.WriteString("```\n")

	case *ast.CodeBlock:
		output.WriteString("```\n")
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			output.Write(line.Value(source))
		}
		output.WriteString("```\n")

	case *ast.List:
		// Lists are handled by their items
		return

	case *ast.ListItem:
		// Only process list items at the top level of the walk
		// (not when we're already inside a list item's paragraph)
		if n.Parent().Kind() == ast.KindList {
			marker := "*"
			if n.Parent().(*ast.List).IsOrdered() {
				marker = "*."
			}

			// Calculate indentation based on nesting level
			depth := 0
			parent := n.Parent()
			for parent != nil {
				if parent.Kind() == ast.KindList {
					depth++
				}
				parent = parent.Parent()
			}
			depth-- // Adjust because we counted the immediate parent

			indent := strings.Repeat("\t", depth)
			output.WriteString(indent + marker + " ")

			// Process list item contents (can be TextBlock, Paragraph, or nested List)
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				if child.Kind() == ast.KindParagraph || child.Kind() == ast.KindTextBlock {
					convertMarkdownInlineChildren(child, source, output, warnings)
				} else if child.Kind() == ast.KindList {
					output.WriteString("\n")
				}
			}
			output.WriteString("\n")
		}

	case *ast.ThematicBreak:
		output.WriteString("----\n")

	case *ast.Blockquote:
		// Convert to mycomarkup quote block
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			output.WriteString("> ")
			if para, ok := child.(*ast.Paragraph); ok {
				convertMarkdownInlineChildren(para, source, output, warnings)
			} else {
				var quoteContent strings.Builder
				convertMarkdownNodeToMycomarkup(child, source, &quoteContent, warnings)
				output.WriteString(strings.TrimSuffix(quoteContent.String(), "\n"))
			}
			output.WriteString("\n")
		}

	case *gast.Table:
		convertGFMTableToMycomarkup(n, source, output, warnings)

	case *ast.Image:
		// Will be handled as inline node
		return

	// Inline nodes are handled by their parents
	case *ast.Text, *ast.String, *ast.Link, *ast.Emphasis, *ast.CodeSpan:
		return

	default:
		// Skip unknown node types silently (they're handled by their parents)
		return
	}
}

func convertMarkdownInlineChildren(node ast.Node, source []byte, output *strings.Builder, warnings *[]string) {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		convertMarkdownInlineNode(child, source, output, warnings)
	}
}

func convertMarkdownInlineNode(node ast.Node, source []byte, output *strings.Builder, warnings *[]string) {
	switch n := node.(type) {
	case *ast.Text:
		segment := n.Segment
		output.Write(segment.Value(source))
		if n.SoftLineBreak() {
			output.WriteString("\n")
		} else if n.HardLineBreak() {
			output.WriteString("\n")
		}

	case *ast.String:
		output.Write(n.Value)

	case *ast.CodeSpan:
		output.WriteString("`")
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			if text, ok := child.(*ast.Text); ok {
				output.Write(text.Segment.Value(source))
			}
		}
		output.WriteString("`")

	case *ast.Emphasis:
		if n.Level == 1 {
			// Italic: *text* -> //text//
			// Check if the first child is also emphasis (for ***text*** case)
			firstChild := n.FirstChild()
			if firstChild != nil && firstChild.Kind() == ast.KindEmphasis {
				childEmph := firstChild.(*ast.Emphasis)
				if childEmph.Level == 2 {
					// ***text*** parsed as Emphasis(1, child=Emphasis(2))
					// Convert to **//text//**
					output.WriteString("**//")
					convertMarkdownInlineChildren(childEmph, source, output, warnings)
					output.WriteString("//**")
					return // Don't process children again
				}
			}
			output.WriteString("//")
			convertMarkdownInlineChildren(n, source, output, warnings)
			output.WriteString("//")
		} else if n.Level == 2 {
			// Bold: **text** -> **text**
			output.WriteString("**")
			convertMarkdownInlineChildren(n, source, output, warnings)
			output.WriteString("**")
		}

	case *ast.Link:
		// [text](url) -> [[url | text]]
		// But handle HTML anchors specially
		dest := string(n.Destination)

		output.WriteString("[[")
		output.WriteString(dest)

		// Check if link has custom text (not just the URL)
		if n.FirstChild() != nil {
			var linkText strings.Builder
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				convertMarkdownInlineNode(child, source, &linkText, warnings)
			}

			// Only add " | text" if text differs from URL
			text := linkText.String()
			if text != dest {
				output.WriteString(" | ")
				output.WriteString(text)
			}
		}
		output.WriteString("]]")

	case *ast.HTMLBlock, *ast.RawHTML:
		// Try to parse common HTML tags and convert to mycomarkup
		// For now, preserve HTML as-is
		return

	case *ast.Image:
		// Convert markdown ![alt](src) to img{} block or inline HTML
		output.WriteString("img { ")
		output.Write(n.Destination)
		if n.Title != nil {
			output.WriteString(" | ")
			output.Write(n.Title)
		} else if n.FirstChild() != nil {
			// Use alt text as description
			var altText strings.Builder
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				if text, ok := child.(*ast.Text); ok {
					altText.Write(text.Segment.Value(source))
				}
			}
			if alt := altText.String(); alt != "" {
				output.WriteString(" | ")
				output.WriteString(alt)
			}
		}
		output.WriteString(" }")

	case *gast.Strikethrough:
		// GFM strikethrough ~~text~~ -> Mycomarkup ~~text~~
		output.WriteString("~~")
		convertMarkdownInlineChildren(n, source, output, warnings)
		output.WriteString("~~")

	default:
		// Try to recurse for unknown inline types
		convertMarkdownInlineChildren(n, source, output, warnings)
	}
}

func convertGFMTableToMycomarkup(table *gast.Table, source []byte, output *strings.Builder, warnings *[]string) {
	output.WriteString("table {\n")

	for row := table.FirstChild(); row != nil; row = row.NextSibling() {
		switch r := row.(type) {
		case *gast.TableRow:
			for cell := r.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if c, ok := cell.(*gast.TableCell); ok {
					// Determine if header
					var cellMarker string
					if c.Alignment == gast.AlignNone {
						cellMarker = ""
					} else {
						cellMarker = "th"
					}

					if cellMarker != "" {
						output.WriteString(cellMarker)
						output.WriteString(" { ")
					}

					// Convert cell contents
					var cellContent strings.Builder
					for child := c.FirstChild(); child != nil; child = child.NextSibling() {
						convertMarkdownInlineNode(child, source, &cellContent, warnings)
					}
					output.WriteString(strings.TrimSpace(cellContent.String()))

					if cellMarker != "" {
						output.WriteString(" }")
					}
					output.WriteString("\n")
				}
			}
		}
	}

	output.WriteString("}\n")
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
