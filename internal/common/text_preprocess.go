package common

import (
	"regexp"
	"strings"

	"github.com/Tencent/WeKnora/internal/types"
)

var (
	reMultiSpace   = regexp.MustCompile(`[^\S\n]{2,}`)
	reMultiNewline = regexp.MustCompile(`\n{3,}`)
)

// ApplyTextPreprocess normalizes text according to knowledge-base ingest settings.
// When no rule is enabled the input is returned unchanged.
func ApplyTextPreprocess(text string, cfg types.TextPreprocessConfig) string {
	if !cfg.Enabled() || text == "" {
		return text
	}
	out := text
	if cfg.NormalizeNewlines {
		out = strings.ReplaceAll(out, "\r\n", "\n")
		out = strings.ReplaceAll(out, "\r", "\n")
		out = reMultiNewline.ReplaceAllString(out, "\n\n")
	}
	if cfg.StripTabs {
		out = strings.ReplaceAll(out, "\t", " ")
	}
	if cfg.TrimLines {
		lines := strings.Split(out, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimSpace(line)
		}
		out = strings.Join(lines, "\n")
	}
	if cfg.CollapseWhitespace {
		out = reMultiSpace.ReplaceAllString(out, " ")
	}
	return out
}
