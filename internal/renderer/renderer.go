package renderer

import (
	"html/template"

	"github.com/bouncepaw/mycorrhiza/internal/hyphae"
	"github.com/bouncepaw/mycorrhiza/internal/mdrenderer"
	"github.com/bouncepaw/mycorrhiza/mycoopts"

	"git.sr.ht/~bouncepaw/mycomarkup/v5"
	"git.sr.ht/~bouncepaw/mycomarkup/v5/mycocontext"
)

// RenderHyphaContent renders hypha text to HTML based on format detected from file path
func RenderHyphaContent(h hyphae.ExistingHypha, content string, hyphaName string) (template.HTML, error) {
	format := hyphae.DetectTextFormat(h.TextFilePath())

	switch format {
	case hyphae.FormatMarkdown:
		html, err := mdrenderer.Render([]byte(content))
		if err != nil {
			return "", err
		}
		return template.HTML(html), nil

	case hyphae.FormatMycomarkup:
		ctx, _ := mycocontext.ContextFromStringInput(content, mycoopts.MarkupOptions(hyphaName))
		html := mycomarkup.BlocksToHTML(ctx, mycomarkup.BlockTree(ctx))
		return template.HTML(html), nil

	default:
		// Fallback to mycomarkup
		ctx, _ := mycocontext.ContextFromStringInput(content, mycoopts.MarkupOptions(hyphaName))
		html := mycomarkup.BlocksToHTML(ctx, mycomarkup.BlockTree(ctx))
		return template.HTML(html), nil
	}
}

// RenderForPreview renders content for preview (when format is known from UI)
func RenderForPreview(content string, format hyphae.TextFormat, hyphaName string) (template.HTML, error) {
	switch format {
	case hyphae.FormatMarkdown:
		html, err := mdrenderer.Render([]byte(content))
		if err != nil {
			return "", err
		}
		return template.HTML(html), nil

	case hyphae.FormatMycomarkup:
		ctx, _ := mycocontext.ContextFromStringInput(content, mycoopts.MarkupOptions(hyphaName))
		html := mycomarkup.BlocksToHTML(ctx, mycomarkup.BlockTree(ctx))
		return template.HTML(html), nil

	default:
		ctx, _ := mycocontext.ContextFromStringInput(content, mycoopts.MarkupOptions(hyphaName))
		html := mycomarkup.BlocksToHTML(ctx, mycomarkup.BlockTree(ctx))
		return template.HTML(html), nil
	}
}
