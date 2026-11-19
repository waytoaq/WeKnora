package httpapi

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Tencent/WeKnora/internal/datasource"
	"github.com/Tencent/WeKnora/internal/types"
)

// Config holds HTTP API connector settings.
// Secrets live in Credentials; non-secret fetch options live in Settings.
// ValidateCredentials may pass settings keys inside Credentials (same pattern as RSS).
type Config struct {
	BaseURL            string
	ListPath           string
	DetailPathTemplate string
	ItemsPath          string
	IDField            string
	TitleField         string
	ContentField       string
	UpdatedAtField     string
	URLField           string
	PageParam          string
	PageSizeParam      string
	PageSize           int
	MaxPages           int
	MaxItems           int
	AuthType           string // bearer | header | none
	Token              string
	AuthHeaders        string // "Name: value" lines
}

func parseConfig(ds *types.DataSourceConfig) (*Config, error) {
	if ds == nil {
		return nil, fmt.Errorf("%w: config is nil", datasource.ErrInvalidConfig)
	}
	cfg := &Config{
		IDField:        "id",
		TitleField:     "title",
		ContentField:   "content",
		UpdatedAtField: "updated_at",
		URLField:       "url",
		PageParam:      "page",
		PageSizeParam:  "page_size",
		PageSize:       50,
		MaxPages:       100,
		MaxItems:       5000,
		AuthType:       "bearer",
	}

	mergeMap := func(m map[string]interface{}) {
		if m == nil {
			return
		}
		if v := strVal(m, "base_url"); v != "" {
			cfg.BaseURL = v
		}
		if v := strVal(m, "list_path"); v != "" {
			cfg.ListPath = v
		}
		if v := strVal(m, "detail_path_template"); v != "" {
			cfg.DetailPathTemplate = v
		}
		if v := strVal(m, "items_path"); v != "" {
			cfg.ItemsPath = v
		}
		if v := strVal(m, "id_field"); v != "" {
			cfg.IDField = v
		}
		if v := strVal(m, "title_field"); v != "" {
			cfg.TitleField = v
		}
		if v := strVal(m, "content_field"); v != "" {
			cfg.ContentField = v
		}
		if v := strVal(m, "updated_at_field"); v != "" {
			cfg.UpdatedAtField = v
		}
		if v := strVal(m, "url_field"); v != "" {
			cfg.URLField = v
		}
		if v := strVal(m, "page_param"); v != "" {
			cfg.PageParam = v
		}
		if v := strVal(m, "page_size_param"); v != "" {
			cfg.PageSizeParam = v
		}
		if v := strVal(m, "auth_type"); v != "" {
			cfg.AuthType = strings.ToLower(v)
		}
		if v := strVal(m, "token"); v != "" {
			cfg.Token = v
		}
		if v := strVal(m, "auth_headers"); v != "" {
			cfg.AuthHeaders = v
		}
		if n := intVal(m, "page_size"); n > 0 {
			cfg.PageSize = n
		}
		if n := intVal(m, "max_pages"); n > 0 {
			cfg.MaxPages = n
		}
		if n := intVal(m, "max_items"); n > 0 {
			cfg.MaxItems = n
		}
	}

	mergeMap(ds.Credentials)
	mergeMap(ds.Settings)

	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	cfg.ListPath = strings.TrimSpace(cfg.ListPath)
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("%w: base_url is required", datasource.ErrInvalidCredentials)
	}
	if cfg.ListPath == "" {
		return nil, fmt.Errorf("%w: list_path is required", datasource.ErrInvalidCredentials)
	}
	if cfg.PageSize > 500 {
		cfg.PageSize = 500
	}
	if cfg.MaxPages > 1000 {
		cfg.MaxPages = 1000
	}
	if cfg.MaxItems > 50000 {
		cfg.MaxItems = 50000
	}
	return cfg, nil
}

func strVal(m map[string]interface{}, key string) string {
	raw, ok := m[key]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return strings.Trim(string(b), `"`)
	}
}

func intVal(m map[string]interface{}, key string) int {
	raw, ok := m[key]
	if !ok || raw == nil {
		return 0
	}
	switch v := raw.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(v))
		return n
	case json.Number:
		n, _ := v.Int64()
		return int(n)
	default:
		return 0
	}
}

type syncCursorState struct {
	LastSyncTime string            `json:"last_sync_time"`
	ItemTimes    map[string]string `json:"item_times"`
}
