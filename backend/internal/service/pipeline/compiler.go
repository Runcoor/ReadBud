// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

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

// StyleConfig captures a complete design system for a single article preset.
// All values are inline-CSS-ready strings so they can be dropped into style="...".
type StyleConfig struct {
	Name        string
	DisplayName string
	Description string

	// Color palette
	Ink        string // primary heading / strong text
	Body       string // paragraph text
	Mute       string // secondary / meta
	Faint      string // tertiary
	Line       string // divider, hairline border
	Paper      string // page background
	PaperAlt   string // panel / lead / summary background
	Accent     string // accent (yellow / red / orange)
	AccentSoft string // accent tint for highlight backgrounds

	// Typography stacks (full font-family declarations)
	SansStack  string
	SerifStack string
	MonoStack  string

	// Decoration policy
	H2Decor   string // "number" | "bar" | "centered-rule"
	LeadDecor string // "box" | "drop-cap" | "lead-text"
	TitleAlign string // "left" | "center"
}

// Three production presets, each modeled on a finalized HTML mockup.
var presetStyles = map[string]StyleConfig{
	"minimal": {
		Name: "minimal", DisplayName: "极简专业", Description: "黑白底色 · 荧光黄高亮 · 衬线标题 · 等宽编号",
		Ink: "#111111", Body: "#2A2A2A", Mute: "#888888", Faint: "#BFBFBF",
		Line: "#ECECEC", Paper: "#FFFFFF", PaperAlt: "#FAFAFA",
		Accent: "#111111", AccentSoft: "#FFE94D",
		SansStack:  `'Noto Sans SC', -apple-system, 'PingFang SC', sans-serif`,
		SerifStack: `'Noto Serif SC', 'Songti SC', serif`,
		MonoStack:  `'JetBrains Mono', 'SF Mono', Menlo, monospace`,
		H2Decor:    "number", LeadDecor: "lead-text", TitleAlign: "left",
	},
	"magazine": {
		Name: "magazine", DisplayName: "杂志编辑", Description: "纸张底色 · 报刊红强调 · Bodoni 大字 · 报头报尾",
		Ink: "#0A0A0A", Body: "#1A1A1A", Mute: "#6B6862", Faint: "#B8B3A8",
		Line: "#1A1A1A", Paper: "#F2EFE8", PaperAlt: "#E8E3D6",
		Accent: "#E63946", AccentSoft: "#F9DDD9",
		SansStack:  `'Noto Sans SC', -apple-system, 'PingFang SC', sans-serif`,
		SerifStack: `'Noto Serif SC', 'Songti SC', serif`,
		MonoStack:  `'JetBrains Mono', 'SF Mono', Menlo, monospace`,
		H2Decor:    "bar", LeadDecor: "drop-cap", TitleAlign: "left",
	},
	"stitch": {
		Name: "stitch", DisplayName: "暖橙手账", Description: "米色底 · 暖橙强调 · 居中标题 · 装饰短横",
		Ink: "#1A1815", Body: "#2C2A26", Mute: "#8A8780", Faint: "#B8B3A8",
		Line: "#E8E4DC", Paper: "#FCFAF5", PaperAlt: "#FDF2E5",
		Accent: "#D2691E", AccentSoft: "#FDF2E5",
		SansStack:  `'Noto Sans SC', -apple-system, 'PingFang SC', sans-serif`,
		SerifStack: `'Noto Serif SC', 'Songti SC', serif`,
		MonoStack:  `'JetBrains Mono', 'SF Mono', Menlo, monospace`,
		H2Decor:    "centered-rule", LeadDecor: "box", TitleAlign: "center",
	},
}

// PresetStyleNames returns the canonical preset identifiers in display order.
func PresetStyleNames() []string {
	return []string{"minimal", "magazine", "stitch"}
}

// GetStyleConfig returns the style config for a given style name, falling back to minimal.
func GetStyleConfig(styleName string) StyleConfig {
	if s, ok := presetStyles[styleName]; ok {
		return s
	}
	return presetStyles["minimal"]
}

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

	sc := GetStyleConfig("minimal")
	compiledHTML := renderArticle(d.Title, blocks, sc)

	d.CompiledHTML = compiledHTML
	if err := s.draftRepo.Update(ctx, d); err != nil {
		return "", fmt.Errorf("compilerService.Compile: save: %w", err)
	}
	return compiledHTML, nil
}

// CompileStyled compiles a draft using the named preset.
func (s *CompilerService) CompileStyled(ctx context.Context, draftID int64, styleName string) (string, error) {
	d, err := s.draftRepo.FindByID(ctx, draftID)
	if err != nil {
		return "", fmt.Errorf("compilerService.CompileStyled: %w", err)
	}
	if d == nil {
		return "", fmt.Errorf("compilerService.CompileStyled: draft %d not found", draftID)
	}
	blocks, err := s.blockRepo.FindByDraftID(ctx, draftID)
	if err != nil {
		return "", fmt.Errorf("compilerService.CompileStyled: %w", err)
	}

	sc := GetStyleConfig(styleName)
	compiledHTML := renderArticle(d.Title, blocks, sc)

	d.CompiledHTML = compiledHTML
	if err := s.draftRepo.Update(ctx, d); err != nil {
		return "", fmt.Errorf("compilerService.CompileStyled: save: %w", err)
	}
	return compiledHTML, nil
}

// renderArticle produces the full <section>…</section> output for a draft.
func renderArticle(title string, blocks []draft.ArticleBlock, sc StyleConfig) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		`<section style="max-width:100%%;margin:0;padding:0;background:%s;color:%s;font-family:%s;font-size:15px;line-height:1.85;">`,
		sc.Paper, sc.Body, sc.SansStack))

	// Article title — only emit if the first block isn't already a title block.
	hasTitleBlock := len(blocks) > 0 && blocks[0].BlockType == draft.BlockTypeTitle
	if !hasTitleBlock && strings.TrimSpace(title) != "" {
		sb.WriteString(renderTitle(title, sc))
	}

	sectionIdx := 0
	for _, b := range blocks {
		if b.Status != "" && b.Status != "active" {
			continue
		}
		if b.BlockType == draft.BlockTypeSection && b.Heading != nil && strings.TrimSpace(*b.Heading) != "" {
			sectionIdx++
		}
		sb.WriteString(CompileBlockStyled(b, sc, sectionIdx))
	}

	sb.WriteString("</section>")
	out := sb.String()
	// WeChat editor caps inline payload at ~20K characters.
	if len(out) > 20000 {
		out = out[:20000] + "</section>"
	}
	return out
}

// CompileBlockStyled renders a single block using the given style config.
// sectionIdx is the 1-based index of section headings encountered so far (used
// for numbered H2 decorations).
func CompileBlockStyled(b draft.ArticleBlock, sc StyleConfig, sectionIdx int) string {
	if b.HTMLFragment != nil && strings.TrimSpace(*b.HTMLFragment) != "" {
		return *b.HTMLFragment + "\n"
	}

	switch b.BlockType {
	case draft.BlockTypeTitle:
		return renderTitle(derefStr(b.Heading), sc)
	case draft.BlockTypeSubtitle:
		return renderSubtitle(derefStr(b.Heading), sc)
	case draft.BlockTypeLead:
		return renderLead(derefStr(b.TextMD), sc)
	case draft.BlockTypeSection:
		return renderSection(derefStr(b.Heading), derefStr(b.TextMD), sectionIdx, sc)
	case draft.BlockTypeImage:
		return renderImage(b, sc)
	case draft.BlockTypeChart:
		return renderChart(b, sc)
	case draft.BlockTypeQuote:
		return renderQuote(derefStr(b.TextMD), sc)
	case draft.BlockTypeChecklist:
		return renderChecklist(derefStr(b.TextMD), sc)
	case draft.BlockTypeSummary:
		return renderSummary(derefStr(b.TextMD), sc)
	case draft.BlockTypeCTA:
		return renderCTA(derefStr(b.TextMD), sc)
	default:
		return renderParagraph(derefStr(b.TextMD), sc)
	}
}

// ---------- Block renderers ----------

func renderTitle(title string, sc StyleConfig) string {
	if strings.TrimSpace(title) == "" {
		return ""
	}
	align := "left"
	if sc.TitleAlign == "center" {
		align = "center"
	}
	switch sc.Name {
	case "magazine":
		return fmt.Sprintf(
			`<h1 style="margin:24px 22px 14px;padding:0;font-family:%s;font-weight:900;font-size:32px;line-height:1.15;letter-spacing:-0.5px;color:%s;text-align:%s;">%s</h1>`+"\n",
			sc.SerifStack, sc.Ink, align, escape(title))
	case "stitch":
		// title with rule underline
		return fmt.Sprintf(
			`<h1 style="margin:24px 20px 14px;padding:0;font-family:%s;font-weight:700;font-size:26px;line-height:1.35;letter-spacing:0.5px;color:%s;text-align:%s;">%s</h1>`+"\n"+
				`<div style="width:60px;height:1px;background:%s;opacity:0.4;margin:0 auto 18px;"></div>`+"\n",
			sc.SerifStack, sc.Accent, align, escape(title), sc.Accent)
	default: // minimal
		return fmt.Sprintf(
			`<h1 style="margin:24px 24px 14px;padding:0;font-family:%s;font-weight:700;font-size:30px;line-height:1.35;letter-spacing:-0.3px;color:%s;text-align:%s;">%s</h1>`+"\n",
			sc.SerifStack, sc.Ink, align, escape(title))
	}
}

func renderSubtitle(text string, sc StyleConfig) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	switch sc.Name {
	case "magazine":
		return fmt.Sprintf(
			`<div style="margin:0 22px 24px;font-family:%s;font-style:italic;font-size:16px;line-height:1.7;color:%s;">%s</div>`+"\n",
			sc.SerifStack, sc.Mute, escape(text))
	case "stitch":
		return fmt.Sprintf(
			`<div style="margin:0 20px 22px;font-family:%s;font-size:14px;line-height:1.7;color:%s;text-align:center;">%s</div>`+"\n",
			sc.SansStack, sc.Mute, escape(text))
	default:
		return fmt.Sprintf(
			`<div style="margin:0 24px 24px;padding-bottom:18px;border-bottom:1px solid %s;font-size:15px;line-height:1.8;color:%s;">%s</div>`+"\n",
			sc.Line, sc.Mute, escape(text))
	}
}

func renderLead(text string, sc StyleConfig) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	body := mdInline(text, sc)
	switch sc.LeadDecor {
	case "drop-cap":
		first, rest := splitFirstRune(text)
		// Use float drop cap — must keep text inline-flow.
		return fmt.Sprintf(
			`<p style="margin:18px 22px 18px;padding:0;font-size:17px;line-height:1.8;color:%s;letter-spacing:0.2px;">`+
				`<span style="float:left;font-family:%s;font-weight:900;font-size:60px;line-height:0.9;margin:6px 8px 0 0;color:%s;">%s</span>`+
				`%s</p>`+"\n",
			sc.Ink, sc.SerifStack, sc.Accent, escape(first), mdInline(rest, sc))
	case "box":
		return fmt.Sprintf(
			`<p style="margin:18px 16px 22px;padding:16px 18px;background:%s;border-radius:8px;font-size:15px;line-height:1.85;color:%s;">%s</p>`+"\n",
			sc.AccentSoft, sc.Ink, body)
	default: // lead-text
		return fmt.Sprintf(
			`<p style="margin:18px 24px 22px;padding-bottom:20px;border-bottom:1px solid %s;font-size:15px;line-height:1.85;color:%s;">%s</p>`+"\n",
			sc.Line, sc.Mute, body)
	}
}

func renderSection(heading, text string, sectionIdx int, sc StyleConfig) string {
	var sb strings.Builder
	if strings.TrimSpace(heading) != "" {
		sb.WriteString(renderH2(heading, sectionIdx, sc))
	}
	if strings.TrimSpace(text) != "" {
		sb.WriteString(mdToHTML(text, sc))
	}
	return sb.String()
}

func renderH2(heading string, sectionIdx int, sc StyleConfig) string {
	switch sc.H2Decor {
	case "number":
		num := fmt.Sprintf("%02d", sectionIdx)
		if sectionIdx <= 0 {
			num = "·"
		}
		return fmt.Sprintf(
			`<h2 style="margin:36px 24px 14px;font-family:%s;font-weight:700;font-size:20px;line-height:1.4;letter-spacing:-0.2px;color:%s;display:flex;align-items:baseline;gap:12px;">`+
				`<span style="font-family:%s;font-size:11px;font-weight:500;color:%s;letter-spacing:1px;flex-shrink:0;">%s</span>`+
				`<span>%s</span></h2>`+"\n",
			sc.SerifStack, sc.Ink, sc.MonoStack, sc.Mute, num, escape(heading))
	case "bar":
		return fmt.Sprintf(
			`<div style="margin:34px 22px 14px;">`+
				`<div style="width:36px;height:4px;background:%s;margin-bottom:12px;"></div>`+
				`<h2 style="margin:0;padding:0;font-family:%s;font-weight:900;font-size:26px;line-height:1.2;letter-spacing:-0.5px;color:%s;">%s</h2>`+
				`</div>`+"\n",
			sc.Accent, sc.SerifStack, sc.Ink, escape(heading))
	case "centered-rule":
		return fmt.Sprintf(
			`<h2 style="margin:36px 20px 8px;font-family:%s;font-weight:700;font-size:22px;line-height:1.4;color:%s;text-align:center;letter-spacing:0.5px;">%s</h2>`+"\n"+
				`<div style="width:56px;height:1px;background:%s;opacity:0.4;margin:0 auto 22px;"></div>`+"\n",
			sc.SerifStack, sc.Accent, escape(heading), sc.Accent)
	default:
		return fmt.Sprintf(
			`<h2 style="margin:32px 24px 12px;font-family:%s;font-weight:700;font-size:20px;line-height:1.4;color:%s;">%s</h2>`+"\n",
			sc.SerifStack, sc.Ink, escape(heading))
	}
}

func renderImage(b draft.ArticleBlock, sc StyleConfig) string {
	// image blocks normally have HTMLFragment populated by the image_match stage.
	// Reaching this branch means the asset is missing — emit a soft placeholder.
	caption := derefStr(b.PromptText)
	return fmt.Sprintf(
		`<figure style="margin:22px 16px;"><div style="aspect-ratio:5/4;background:%s;border:1px solid %s;border-radius:8px;"></div>`+
			`<figcaption style="margin-top:10px;font-size:12px;color:%s;text-align:center;letter-spacing:0.3px;">%s</figcaption></figure>`+"\n",
		sc.PaperAlt, sc.Line, sc.Mute, escape(caption))
}

func renderChart(b draft.ArticleBlock, sc StyleConfig) string {
	if b.HTMLFragment != nil && *b.HTMLFragment != "" {
		return *b.HTMLFragment + "\n"
	}
	caption := derefStr(b.PromptText)
	return fmt.Sprintf(
		`<p style="margin:18px 24px;text-align:center;font-family:%s;font-size:12px;color:%s;letter-spacing:0.5px;">[图表] %s</p>`+"\n",
		sc.MonoStack, sc.Mute, escape(caption))
}

func renderQuote(text string, sc StyleConfig) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	switch sc.Name {
	case "magazine":
		return fmt.Sprintf(
			`<blockquote style="margin:24px 14px;padding:22px 24px 22px 52px;background:%s;color:%s;font-family:%s;font-weight:600;font-size:18px;line-height:1.6;letter-spacing:0.3px;position:relative;">`+
				`<span style="position:absolute;left:14px;top:-6px;font-family:%s;font-size:64px;color:%s;line-height:1;">&ldquo;</span>%s</blockquote>`+"\n",
			sc.Ink, sc.Paper, sc.SerifStack, sc.SerifStack, sc.Accent, escape(text))
	case "stitch":
		return fmt.Sprintf(
			`<blockquote style="margin:20px 12px;padding:16px 18px;background:%s;border-left:3px solid %s;border-radius:6px;font-size:15px;line-height:1.85;color:%s;">%s</blockquote>`+"\n",
			sc.AccentSoft, sc.Accent, sc.Ink, escape(text))
	default: // minimal
		return fmt.Sprintf(
			`<blockquote style="margin:22px 24px;padding:14px 0 14px 18px;border-left:2px solid %s;font-size:15px;line-height:1.85;color:%s;">%s</blockquote>`+"\n",
			sc.Ink, sc.Body, escape(text))
	}
}

func renderChecklist(text string, sc StyleConfig) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	var sb strings.Builder
	switch sc.Name {
	case "minimal":
		sb.WriteString(fmt.Sprintf(`<div style="margin:20px 24px;padding:0;">`))
		for _, raw := range lines {
			line := stripBullet(strings.TrimSpace(raw))
			if line == "" {
				continue
			}
			sb.WriteString(fmt.Sprintf(
				`<div style="position:relative;padding:10px 0 10px 24px;border-bottom:1px solid %s;font-size:14px;line-height:1.75;color:%s;">`+
					`<span style="position:absolute;left:6px;top:18px;width:6px;height:6px;border-radius:50%%;background:%s;"></span>%s</div>`,
				sc.Line, sc.Body, sc.Ink, mdInline(line, sc)))
		}
		sb.WriteString(`</div>`)
	case "magazine":
		sb.WriteString(fmt.Sprintf(`<div style="margin:22px 14px;padding:16px 20px;border:1px solid %s;background:%s;">`, sc.Ink, sc.Paper))
		for _, raw := range lines {
			line := stripBullet(strings.TrimSpace(raw))
			if line == "" {
				continue
			}
			sb.WriteString(fmt.Sprintf(
				`<div style="padding:6px 0 6px 18px;position:relative;font-size:14px;line-height:1.75;color:%s;">`+
					`<span style="position:absolute;left:0;top:6px;color:%s;font-weight:700;">→</span>%s</div>`,
				sc.Body, sc.Accent, mdInline(line, sc)))
		}
		sb.WriteString(`</div>`)
	default: // stitch
		sb.WriteString(fmt.Sprintf(`<div style="margin:20px 16px;">`))
		for _, raw := range lines {
			line := stripBullet(strings.TrimSpace(raw))
			if line == "" {
				continue
			}
			sb.WriteString(fmt.Sprintf(
				`<div style="padding:12px 0;border-bottom:1px solid %s;font-size:15px;line-height:1.75;color:%s;">`+
					`<span style="display:inline-block;width:18px;color:%s;font-weight:700;">·</span>%s</div>`,
				sc.Line, sc.Body, sc.Accent, mdInline(line, sc)))
		}
		sb.WriteString(`</div>`)
	}
	sb.WriteString("\n")
	return sb.String()
}

func renderSummary(text string, sc StyleConfig) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	body := mdInline(text, sc)
	switch sc.Name {
	case "magazine":
		return fmt.Sprintf(
			`<div style="margin:30px 14px 22px;padding:24px 22px;background:%s;color:%s;">`+
				`<div style="font-family:%s;font-size:10px;letter-spacing:3px;color:%s;margin-bottom:12px;text-transform:uppercase;">— FINAL TAKE</div>`+
				`<div style="font-family:%s;font-weight:700;font-size:18px;line-height:1.5;color:%s;">%s</div></div>`+"\n",
			sc.Ink, sc.Paper, sc.MonoStack, sc.Accent, sc.SerifStack, sc.Paper, body)
	case "stitch":
		return fmt.Sprintf(
			`<div style="margin:24px 16px;padding:20px;background:%s;border-radius:8px;border:1px solid %s;font-size:15px;line-height:1.8;color:%s;">`+
				`<div style="font-family:%s;font-size:10px;letter-spacing:2px;color:%s;margin-bottom:8px;text-transform:uppercase;">总结</div>`+
				`%s</div>`+"\n",
			sc.AccentSoft, sc.Line, sc.Ink, sc.MonoStack, sc.Accent, body)
	default: // minimal
		return fmt.Sprintf(
			`<div style="margin:28px 24px;padding:20px;background:%s;border-radius:8px;font-size:15px;line-height:1.85;color:%s;">`+
				`<div style="font-size:14px;color:%s;margin-bottom:8px;">✦</div>%s</div>`+"\n",
			sc.PaperAlt, sc.Ink, sc.Mute, body)
	}
}

func renderCTA(text string, sc StyleConfig) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	lines := splitLines(text)
	var sb strings.Builder
	switch sc.Name {
	case "magazine":
		sb.WriteString(fmt.Sprintf(
			`<div style="margin:30px 22px 0;padding:24px 0 18px;border-top:3px double %s;border-bottom:1px solid %s;">`+
				`<div style="font-family:%s;font-size:10px;letter-spacing:2px;color:%s;margin-bottom:12px;text-transform:uppercase;">— EDITOR'S NOTE</div>`,
			sc.Ink, sc.Ink, sc.MonoStack, sc.Accent))
		for i, l := range lines {
			border := fmt.Sprintf(`border-bottom:1px dashed %s;`, sc.Faint)
			if i == len(lines)-1 {
				border = ""
			}
			sb.WriteString(fmt.Sprintf(
				`<div style="position:relative;padding:8px 0 8px 22px;font-size:13px;line-height:1.75;color:%s;%s">`+
					`<span style="position:absolute;left:0;top:8px;color:%s;font-weight:700;">→</span>%s</div>`,
				sc.Body, border, sc.Accent, mdInline(l, sc)))
		}
		sb.WriteString(`</div>`)
	case "stitch":
		sb.WriteString(fmt.Sprintf(
			`<div style="margin:30px 16px 0;padding:0;">`+
				`<div style="height:1px;background:%s;margin-bottom:24px;position:relative;"><span style="position:absolute;left:50%%;top:50%%;transform:translate(-50%%,-50%%);background:%s;padding:0 10px;color:%s;font-size:10px;">◆</span></div>`,
			sc.Line, sc.Paper, sc.Accent))
		for _, l := range lines {
			sb.WriteString(fmt.Sprintf(
				`<p style="margin:0 0 12px 0;font-size:15px;line-height:1.85;color:%s;font-weight:600;">%s</p>`,
				sc.Accent, mdInline(l, sc)))
		}
		sb.WriteString(`</div>`)
	default: // minimal
		sb.WriteString(fmt.Sprintf(
			`<div style="margin:32px 24px 0;padding:24px;background:%s;border-radius:12px;">`+
				`<div style="font-family:%s;font-weight:600;font-size:16px;color:%s;margin-bottom:14px;">如果觉得有用</div>`,
			sc.PaperAlt, sc.SerifStack, sc.Ink))
		for _, l := range lines {
			sb.WriteString(fmt.Sprintf(
				`<div style="padding:8px 0;font-size:13.5px;line-height:1.7;color:%s;display:flex;gap:10px;">`+
					`<span style="color:%s;flex-shrink:0;">→</span><span>%s</span></div>`,
				sc.Body, sc.Mute, mdInline(l, sc)))
		}
		sb.WriteString(`</div>`)
	}
	sb.WriteString("\n")
	return sb.String()
}

func renderParagraph(text string, sc StyleConfig) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	return mdToHTML(text, sc)
}

// ---------- Lightweight Markdown → inline-styled HTML ----------

// mdToHTML renders a multi-line markdown chunk into a sequence of inline-styled
// block elements (p / h2 / h3 / ul / blockquote / hr).
func mdToHTML(md string, sc StyleConfig) string {
	if strings.TrimSpace(md) == "" {
		return ""
	}
	var sb strings.Builder
	for _, line := range strings.Split(md, "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		switch {
		case strings.HasPrefix(t, "### "):
			sb.WriteString(fmt.Sprintf(
				`<h3 style="margin:24px 24px 10px;font-family:%s;font-weight:700;font-size:16px;color:%s;">%s</h3>`+"\n",
				sc.SerifStack, sc.Ink, escape(t[4:])))
		case strings.HasPrefix(t, "## "):
			// Emit as styled paragraph-heading; section H2 should use renderH2 directly.
			sb.WriteString(fmt.Sprintf(
				`<h3 style="margin:24px 24px 10px;font-family:%s;font-weight:700;font-size:17px;color:%s;">%s</h3>`+"\n",
				sc.SerifStack, sc.Ink, escape(t[3:])))
		case strings.HasPrefix(t, "# "):
			sb.WriteString(fmt.Sprintf(
				`<h3 style="margin:24px 24px 10px;font-family:%s;font-weight:700;font-size:18px;color:%s;">%s</h3>`+"\n",
				sc.SerifStack, sc.Ink, escape(t[2:])))
		case strings.HasPrefix(t, "- "), strings.HasPrefix(t, "* "):
			item := stripBullet(t)
			sb.WriteString(fmt.Sprintf(
				`<p style="margin:0 24px 8px;padding-left:14px;position:relative;font-size:15px;line-height:1.85;color:%s;">`+
					`<span style="position:absolute;left:0;top:0;color:%s;">•</span>%s</p>`+"\n",
				sc.Body, sc.Accent, mdInline(item, sc)))
		case strings.HasPrefix(t, "> "):
			sb.WriteString(renderQuote(t[2:], sc))
		case t == "---" || t == "***":
			sb.WriteString(fmt.Sprintf(
				`<hr style="border:none;border-top:1px solid %s;margin:28px 24px;"/>`+"\n", sc.Line))
		default:
			sb.WriteString(fmt.Sprintf(
				`<p style="margin:0 24px 18px;font-size:15px;line-height:1.85;color:%s;letter-spacing:0.2px;">%s</p>`+"\n",
				sc.Body, mdInline(t, sc)))
		}
	}
	return sb.String()
}

// mdInline applies inline markdown (bold, code, highlight) to a single line and
// HTML-escapes the rest. Patterns: **bold**, `code`, ==highlight==.
func mdInline(text string, sc StyleConfig) string {
	if text == "" {
		return ""
	}
	out := escape(text)
	// ==highlight== (yellow span / accent underline)
	out = replacePairs(out, "==", "==", func(inner string) string {
		switch sc.Name {
		case "minimal":
			return fmt.Sprintf(
				`<span style="background:linear-gradient(transparent 60%%,%s 60%%);padding:0 2px;">%s</span>`,
				sc.AccentSoft, inner)
		case "magazine":
			return fmt.Sprintf(`<em style="font-style:italic;color:%s;font-weight:700;">%s</em>`, sc.Accent, inner)
		default:
			return fmt.Sprintf(`<span style="color:%s;font-weight:600;">%s</span>`, sc.Accent, inner)
		}
	})
	// **bold**
	out = replacePairs(out, "**", "**", func(inner string) string {
		return fmt.Sprintf(`<strong style="color:%s;font-weight:700;">%s</strong>`, sc.Ink, inner)
	})
	// `code`
	out = replacePairs(out, "`", "`", func(inner string) string {
		return fmt.Sprintf(
			`<code style="background:%s;color:%s;padding:1px 6px;border-radius:3px;font-family:%s;font-size:13px;">%s</code>`,
			sc.PaperAlt, sc.Ink, sc.MonoStack, inner)
	})
	return out
}

// ---------- Helpers ----------

func escape(s string) string { return html.EscapeString(s) }

// replacePairs scans s for matching open/close delimiters and rewrites each pair
// using fn(inner). Operates on already-escaped HTML (so delimiters must be ASCII).
func replacePairs(s, open, close string, fn func(string) string) string {
	if open == "" || close == "" {
		return s
	}
	var sb strings.Builder
	for {
		i := strings.Index(s, open)
		if i < 0 {
			sb.WriteString(s)
			break
		}
		sb.WriteString(s[:i])
		rest := s[i+len(open):]
		j := strings.Index(rest, close)
		if j < 0 {
			sb.WriteString(s[i:])
			break
		}
		sb.WriteString(fn(rest[:j]))
		s = rest[j+len(close):]
	}
	return sb.String()
}

func splitLines(text string) []string {
	out := []string{}
	for _, raw := range strings.Split(text, "\n") {
		t := stripBullet(strings.TrimSpace(raw))
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func stripBullet(s string) string {
	s = strings.TrimSpace(s)
	for _, p := range []string{"- ", "* ", "• ", "→ "} {
		if strings.HasPrefix(s, p) {
			return strings.TrimSpace(s[len(p):])
		}
	}
	return s
}

func splitFirstRune(s string) (string, string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ""
	}
	for i, r := range s {
		_ = r
		if i == 0 {
			continue
		}
		return s[:i], s[i:]
	}
	return s, ""
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
