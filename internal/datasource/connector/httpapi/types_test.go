package httpapi

import (
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
)

func TestParseConfigRequiresBaseAndList(t *testing.T) {
	_, err := parseConfig(&types.DataSourceConfig{
		Settings: map[string]interface{}{"base_url": "https://example.com"},
	})
	if err == nil {
		t.Fatal("list_path required")
	}
	cfg, err := parseConfig(&types.DataSourceConfig{
		Settings: map[string]interface{}{
			"base_url":  "https://example.com/",
			"list_path": "/api/items",
			"page_size": 20,
		},
		Credentials: map[string]interface{}{
			"token": "secret",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "https://example.com" {
		t.Fatalf("base url trim: %q", cfg.BaseURL)
	}
	if cfg.PageSize != 20 || cfg.IDField != "id" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}
