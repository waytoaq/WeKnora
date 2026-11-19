package common

import (
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
)

func TestApplyTextPreprocess(t *testing.T) {
	in := "a\t\tb  c\r\n\r\n\r\n d "
	out := ApplyTextPreprocess(in, types.TextPreprocessConfig{
		CollapseWhitespace: true,
		NormalizeNewlines:  true,
		StripTabs:          true,
		TrimLines:          true,
	})
	want := "a b c\n\nd"
	if out != want {
		t.Fatalf("got %q want %q", out, want)
	}
	if ApplyTextPreprocess(in, types.TextPreprocessConfig{}) != in {
		t.Fatalf("disabled config must be a no-op")
	}
}
