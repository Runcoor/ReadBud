package pipeline

import (
	"context"
	"fmt"
	"html"
	"strings"

	"readbud/internal/domain/draft"
	"readbud/internal/repository/postgres"
)

// CompilerService compiles article blocks into WeChat-compatible HTML.
type CompilerService struct {
	draftRepo postgres.ArticleDraftRepository
	blockRepo postgres.ArticleBlockRepository
}

// NewCompilerService creates a new CompilerService.
func NewCompilerService(draftRepo postgres.ArticleDraftRepository, blockRepo postgres.ArticleBlockRepository) *CompilerService {
	return &CompilerService{draftRepo: draftRepo, blockRepo: blockRepo}
}

// WeChat HTML style constants — all inline per WeChat requirements.
const (
	styleSection  = `style="margin: 0 0 20px 0; padding: 0; line-height: 1.75; font-size: 16px; color: #333333;"`
	styleH2       = `style="margin: 24px 0 12px 0; font-size: 20px; font-weight: bold; color: #1B3A5C; line-height: 1.4;"`
	styleH3       = `style="margin: 20px 0 10px 0; font-size: 18px; font-weight: bold; color: #333333; line-height: 1.4;"`
	styleP        = `style="margin: 0 0 16px 0; line-height: 1.75; font-size: 16px; color: #4A4A4A;"`
	styleImg      = `style="max-width: 100%; height: auto; display: block; margin: 16px auto; border-radius: 4px;"`
	styleCaption  = `style="text-align: center; font-size: 13px; color: #999; margin-top: 4px; margin-bottom: 16px;"`
	styleLead     = `style="margin: 0 0 24px 0; padding: 16px; background-color: #F5F7FA; border-left: 4px solid #5B8DEF; font-size: 15px; color: #4A4A4A; line-height: 1.75;"`
	styleQuote    = `style="margin: 16px 0; padding: 12px 16px; background-color: #F9FAFB; border-left: 3px solid #D1D5DB; font-size: 15px; color: #666; line-height: 1.6; font-style: italic;"`
	styleCTA      = `style="margin: 24px 0 0 0; padding: 20px; background-color: #F5F7FA; border-radius: 8px; text-align: center; font-size: 15px; color: #1B3A5C; line-height: 1.6;"`
	styleSummary  = `style="margin: 24px 0 16px 0; padding: 16px; background-color: #EEF3FD; border-radius: 6px; font-size: 15px; color: #333; line-height: 1.75;"`
	styleDivider  = `style="border: none; border-top: 1px solid #E5E7EB; margin: 24px 0;"`
)

// Compile compiles an article draft's blocks into WeChat-compatible HTML.
func (s *CompilerService) Compile(ctx context.Context, draftID int64) (string, error) {
	d, err := s.draftRepo.FindByID(ctx, draftID)
	if err != nil {
		return "", fmt.Errorf("compilerService.Compile: %w", err)
	}
	if d == nil {
		return "", fmt.Errorf("compilerService.Compile: draft %d not found", draftID)
	}

	blocks, err := s.blockRepo.FindByDraftID(ctx, draftID)
	if err != nil {
		return "", fmt.Errorf("compilerService.Compile: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(`<section style="max-width: 100%; padding: 0; margin: 0; font-family: -apple-system, BlinkMacSystemFont, 'PingFang SC', 'Microsoft YaHei', sans-serif;">`)
	sb.WriteString("\n")

	for _, block := range blocks {
		if block.Status != "active" {
			continue
		}
		sb.WriteString(compileBlock(block))
		sb.WriteString("\n")
	}

	sb.WriteString("</section>")

	compiledHTML := sb.String()

	// Enforce WeChat 20K char limit
	if len(compiledHTML) > 20000 {
		compiledHTML = compiledHTML[:20000] + "</section>"
	}

	// Store compiled HTML
	d.CompiledHTML = compiledHTML
	if err := s.draftRepo.Update(ctx, d); err != nil {
		return "", fmt.Errorf("compilerService.Compile: save: %w", err)
	}

	return compiledHTML, nil
}

func compileBlock(b draft.ArticleBlock) string {
	switch b.BlockType {
	case draft.BlockTypeTitle:
		title := ""
		if b.Heading != nil {
			title = html.EscapeString(*b.Heading)
		}
		return fmt.Sprintf(`<h1 style="margin: 0 0 8px 0; font-size: 24px; font-weight: bold; color: #1B3A5C; line-height: 1.3;">%s</h1>`, title)

	case draft.BlockTypeSubtitle:
		sub := ""
		if b.Heading != nil {
			sub = html.EscapeString(*b.Heading)
		}
		return fmt.Sprintf(`<h2 %s>%s</h2>`, styleH3, sub)

	case draft.BlockTypeLead:
		text := derefStr(b.TextMD)
		return fmt.Sprintf(`<div %s>%s</div>`, styleLead, markdownToHTML(text))

	case draft.BlockTypeSection:
		var sb strings.Builder
		if b.Heading != nil && *b.Heading != "" {
			sb.WriteString(fmt.Sprintf(`<h2 %s>%s</h2>`, styleH2, html.EscapeString(*b.Heading)))
			sb.WriteString("\n")
		}
		text := derefStr(b.TextMD)
		sb.WriteString(markdownToHTML(text))
		return sb.String()

	case draft.BlockTypeImage:
		// Image blocks reference an asset; if no HTML fragment yet, show placeholder
		if b.HTMLFragment != nil && *b.HTMLFragment != "" {
			return *b.HTMLFragment
		}
		alt := derefStr(b.PromptText)
		return fmt.Sprintf(`<p style="text-align: center;"><img src="" alt="%s" %s/></p>`, html.EscapeString(alt), styleImg)

	case draft.BlockTypeChart:
		if b.HTMLFragment != nil && *b.HTMLFragment != "" {
			return *b.HTMLFragment
		}
		return fmt.Sprintf(`<p %s>[图表: %s]</p>`, styleCaption, html.EscapeString(derefStr(b.PromptText)))

	case draft.BlockTypeQuote:
		text := derefStr(b.TextMD)
		return fmt.Sprintf(`<blockquote %s>%s</blockquote>`, styleQuote, html.EscapeString(text))

	case draft.BlockTypeSummary:
		text := derefStr(b.TextMD)
		return fmt.Sprintf(`<div %s><strong>总结</strong><br/>%s</div>`, styleSummary, markdownToHTML(text))

	case draft.BlockTypeCTA:
		text := derefStr(b.TextMD)
		return fmt.Sprintf(`<div %s>%s</div>`, styleCTA, markdownToHTML(text))

	default:
		text := derefStr(b.TextMD)
		if text != "" {
			return fmt.Sprintf(`<p %s>%s</p>`, styleP, html.EscapeString(text))
		}
		return ""
	}
}

// markdownToHTML is a lightweight Markdown to inline-styled HTML converter.
// For production, consider using goldmark or blackfriday with custom renderers.
func markdownToHTML(md string) string {
	if md == "" {
		return ""
	}

	var sb strings.Builder
	lines := strings.Split(md, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "### ") {
			sb.WriteString(fmt.Sprintf(`<h3 %s>%s</h3>`, styleH3, html.EscapeString(trimmed[4:])))
		} else if strings.HasPrefix(trimmed, "## ") {
			sb.WriteString(fmt.Sprintf(`<h2 %s>%s</h2>`, styleH2, html.EscapeString(trimmed[3:])))
		} else if strings.HasPrefix(trimmed, "# ") {
			// Skip h1 in body (already handled by title block)
			sb.WriteString(fmt.Sprintf(`<h2 %s>%s</h2>`, styleH2, html.EscapeString(trimmed[2:])))
		} else if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			sb.WriteString(fmt.Sprintf(`<p %s>• %s</p>`, styleP, html.EscapeString(trimmed[2:])))
		} else if strings.HasPrefix(trimmed, "> ") {
			sb.WriteString(fmt.Sprintf(`<blockquote %s>%s</blockquote>`, styleQuote, html.EscapeString(trimmed[2:])))
		} else if trimmed == "---" || trimmed == "***" {
			sb.WriteString(fmt.Sprintf(`<hr %s/>`, styleDivider))
		} else {
			sb.WriteString(fmt.Sprintf(`<p %s>%s</p>`, styleP, html.EscapeString(trimmed)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
