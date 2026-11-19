package service

import (
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
)

func TestCompileExtractHintsFromBundle(t *testing.T) {
	hints := CompileExtractHintsFromBundle(&types.GraphSchemaBundle{
		Entities: []*types.KBGraphEntity{
			{Name: "人物", Code: "person", Category: "person", Description: "人物实体"},
		},
		Relations: []*types.KBGraphRelation{
			{Name: "任职于", Code: "works_at", Description: "工作关系"},
		},
		Attributes: []*types.KBGraphAttribute{
			{Name: "职级", Code: "title", ValueType: "free"},
		},
	})
	if len(hints.Tags) != 1 || hints.Tags[0] != "任职于" {
		t.Fatalf("tags = %#v", hints.Tags)
	}
	if len(hints.EntityCodes) != 1 || hints.EntityCodes[0] != "person" {
		t.Fatalf("entity codes = %#v", hints.EntityCodes)
	}
	if !strings.Contains(hints.CustomInstructions, "Allowed entity types") {
		t.Fatalf("missing entity section: %s", hints.CustomInstructions)
	}
	if !strings.Contains(hints.CustomInstructions, "任职于") {
		t.Fatalf("missing relation in instructions")
	}
}

func TestCompileExtractHintsEmpty(t *testing.T) {
	hints := CompileExtractHintsFromBundle(nil)
	if hints == nil || hints.CustomInstructions != "" || len(hints.Tags) != 0 {
		t.Fatalf("expected empty hints, got %#v", hints)
	}
}
