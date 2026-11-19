package sqlconn

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Tencent/WeKnora/internal/datasource"
	"github.com/Tencent/WeKnora/internal/types"
)

// Config for the SQL knowledge connector.
type Config struct {
	Driver            string // postgres | mysql
	Host              string
	Port              int
	Database          string
	User              string
	Password          string
	SSLMode           string
	Query             string
	Table             string
	Where             string
	IDColumn          string
	TitleColumn       string
	ContentColumns    []string
	ContentTemplate   string
	IncrementalColumn string
	PageSize          int
	MaxRows           int
	QueryTimeoutSec   int
}

func parseConfig(ds *types.DataSourceConfig) (*Config, error) {
	if ds == nil {
		return nil, fmt.Errorf("%w: config is nil", datasource.ErrInvalidConfig)
	}
	cfg := &Config{
		Driver:          "postgres",
		Port:            5432,
		SSLMode:         "disable",
		IDColumn:        "id",
		TitleColumn:     "title",
		PageSize:        200,
		MaxRows:         10000,
		QueryTimeoutSec: 60,
	}
	merge := func(m map[string]interface{}) {
		if m == nil {
			return
		}
		if v := strVal(m, "driver"); v != "" {
			cfg.Driver = strings.ToLower(v)
		}
		if v := strVal(m, "host"); v != "" {
			cfg.Host = v
		}
		if n := intVal(m, "port"); n > 0 {
			cfg.Port = n
		}
		if v := strVal(m, "database"); v != "" {
			cfg.Database = v
		}
		if v := strVal(m, "user"); v != "" {
			cfg.User = v
		}
		if v := strVal(m, "password"); v != "" {
			cfg.Password = v
		}
		if v := strVal(m, "ssl_mode"); v != "" {
			cfg.SSLMode = v
		}
		if v := strVal(m, "query"); v != "" {
			cfg.Query = v
		}
		if v := strVal(m, "table"); v != "" {
			cfg.Table = v
		}
		if v := strVal(m, "where"); v != "" {
			cfg.Where = v
		}
		if v := strVal(m, "id_column"); v != "" {
			cfg.IDColumn = v
		}
		if v := strVal(m, "title_column"); v != "" {
			cfg.TitleColumn = v
		}
		if v := strVal(m, "content_template"); v != "" {
			cfg.ContentTemplate = v
		}
		if v := strVal(m, "incremental_column"); v != "" {
			cfg.IncrementalColumn = v
		}
		if cols := strVal(m, "content_columns"); cols != "" {
			cfg.ContentColumns = splitCSV(cols)
		}
		if raw, ok := m["content_columns"]; ok {
			switch t := raw.(type) {
			case []any:
				out := make([]string, 0, len(t))
				for _, item := range t {
					if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
						out = append(out, strings.TrimSpace(s))
					}
				}
				if len(out) > 0 {
					cfg.ContentColumns = out
				}
			case []string:
				if len(t) > 0 {
					cfg.ContentColumns = t
				}
			}
		}
		if n := intVal(m, "page_size"); n > 0 {
			cfg.PageSize = n
		}
		if n := intVal(m, "max_rows"); n > 0 {
			cfg.MaxRows = n
		}
		if n := intVal(m, "query_timeout_sec"); n > 0 {
			cfg.QueryTimeoutSec = n
		}
	}
	merge(ds.Credentials)
	merge(ds.Settings)

	switch cfg.Driver {
	case "postgres", "postgresql", "pg":
		cfg.Driver = "postgres"
		if cfg.Port == 0 {
			cfg.Port = 5432
		}
	case "mysql", "mariadb":
		cfg.Driver = "mysql"
		if cfg.Port == 0 {
			cfg.Port = 3306
		}
	default:
		return nil, fmt.Errorf("%w: unsupported driver %q (postgres|mysql)", datasource.ErrInvalidCredentials, cfg.Driver)
	}
	if strings.TrimSpace(cfg.Host) == "" || strings.TrimSpace(cfg.Database) == "" || strings.TrimSpace(cfg.User) == "" {
		return nil, fmt.Errorf("%w: host, database and user are required", datasource.ErrInvalidCredentials)
	}
	cfg.Query = strings.TrimSpace(cfg.Query)
	cfg.Table = strings.TrimSpace(cfg.Table)
	if cfg.Query == "" && cfg.Table == "" {
		return nil, fmt.Errorf("%w: query or table is required", datasource.ErrInvalidCredentials)
	}
	if cfg.Query != "" {
		if err := ensureReadOnlySelect(cfg.Query); err != nil {
			return nil, err
		}
	}
	if cfg.Table != "" {
		if err := ensureSafeIdent(cfg.Table); err != nil {
			return nil, fmt.Errorf("%w: invalid table: %v", datasource.ErrInvalidConfig, err)
		}
	}
	if cfg.PageSize > 1000 {
		cfg.PageSize = 1000
	}
	if cfg.MaxRows > 100000 {
		cfg.MaxRows = 100000
	}
	return cfg, nil
}

func ensureReadOnlySelect(query string) error {
	q := strings.TrimSpace(query)
	lower := strings.ToLower(q)
	if strings.Contains(q, ";") {
		return fmt.Errorf("%w: multiple statements are not allowed", datasource.ErrInvalidConfig)
	}
	if !strings.HasPrefix(lower, "select") && !strings.HasPrefix(lower, "with") {
		return fmt.Errorf("%w: only SELECT/CTE queries are allowed", datasource.ErrInvalidConfig)
	}
	forbidden := []string{" insert ", " update ", " delete ", " drop ", " alter ", " truncate ", " create ", " grant ", " revoke ", " copy ", " call ", " execute "}
	padded := " " + lower + " "
	for _, f := range forbidden {
		if strings.Contains(padded, f) {
			return fmt.Errorf("%w: query contains forbidden keyword", datasource.ErrInvalidConfig)
		}
	}
	return nil
}

func ensureSafeIdent(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("empty identifier")
	}
	for _, part := range strings.Split(name, ".") {
		part = strings.Trim(part, `"`)
		if part == "" {
			return fmt.Errorf("empty identifier part")
		}
		for i, r := range part {
			ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || (i > 0 && r >= '0' && r <= '9')
			if !ok {
				return fmt.Errorf("unsafe identifier %q", name)
			}
		}
	}
	return nil
}

func quoteIdent(driver, name string) string {
	parts := strings.Split(name, ".")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.Trim(p, `"`)
		if driver == "mysql" {
			out = append(out, "`"+strings.ReplaceAll(p, "`", "``")+"`")
		} else {
			out = append(out, `"`+strings.ReplaceAll(p, `"`, `""`)+`"`)
		}
	}
	return strings.Join(out, ".")
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
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
	LastSyncTime string `json:"last_sync_time"`
	LastValue    string `json:"last_value"`
}
