package sqlconn

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/WeKnora/internal/datasource"
	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/types"
)

var _ datasource.Connector = (*Connector)(nil)

// Connector syncs relational database rows into knowledge documents.
type Connector struct{}

// NewConnector creates a SQL connector.
func NewConnector() *Connector { return &Connector{} }

// Type returns the connector type identifier.
func (c *Connector) Type() string { return types.ConnectorTypeSQL }

// Validate opens a connection and runs a LIMIT 1 probe.
func (c *Connector) Validate(ctx context.Context, config *types.DataSourceConfig) error {
	cfg, err := parseConfig(config)
	if err != nil {
		return err
	}
	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	base, err := buildBaseQuery(cfg)
	if err != nil {
		return err
	}
	rows, err := queryPage(ctx, db, cfg, base, "", 0)
	if err != nil {
		return fmt.Errorf("probe query: %w", err)
	}
	_ = rows
	return nil
}

// ListResources returns a single resource for the configured query/table.
func (c *Connector) ListResources(
	ctx context.Context, config *types.DataSourceConfig, parentID string,
) ([]types.Resource, error) {
	if parentID != "" {
		return []types.Resource{}, nil
	}
	cfg, err := parseConfig(config)
	if err != nil {
		return nil, err
	}
	name := cfg.Table
	if name == "" {
		name = "custom_query"
	}
	desc := cfg.Query
	if desc == "" {
		desc = "SELECT * FROM " + cfg.Table
	}
	if len(desc) > 120 {
		desc = desc[:120] + "…"
	}
	return []types.Resource{{
		ExternalID:  "sql:default",
		Name:        name,
		Type:        "table",
		Description: desc,
	}}, nil
}

// ResolveResourceAncestors is unused for the flat SQL resource.
func (c *Connector) ResolveResourceAncestors(
	ctx context.Context, config *types.DataSourceConfig, resourceIDs []string,
) ([]string, error) {
	return []string{}, nil
}

// FetchAll loads all rows up to MaxRows.
func (c *Connector) FetchAll(
	ctx context.Context, config *types.DataSourceConfig, resourceIDs []string,
) ([]types.FetchedItem, error) {
	items, _, err := c.walk(ctx, config, nil, false)
	return items, err
}

// FetchIncremental loads rows newer than the stored incremental cursor value.
func (c *Connector) FetchIncremental(
	ctx context.Context, config *types.DataSourceConfig, cursor *types.SyncCursor,
) ([]types.FetchedItem, *types.SyncCursor, error) {
	prev := decodeCursor(cursor)
	items, next, err := c.walk(ctx, config, prev, true)
	if err != nil && next == nil {
		return nil, nil, err
	}
	return items, encodeCursor(next), err
}

func (c *Connector) walk(
	ctx context.Context, config *types.DataSourceConfig, prev *syncCursorState, incremental bool,
) ([]types.FetchedItem, *syncCursorState, error) {
	cfg, err := parseConfig(config)
	if err != nil {
		return nil, nil, err
	}
	db, err := openDB(cfg)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	base, err := buildBaseQuery(cfg)
	if err != nil {
		return nil, nil, err
	}

	after := ""
	if incremental && cfg.IncrementalColumn != "" && prev != nil {
		after = prev.LastValue
	}

	next := &syncCursorState{
		LastSyncTime: time.Now().UTC().Format(time.RFC3339),
		LastValue:    after,
	}
	var out []types.FetchedItem
	offset := 0
	if after != "" {
		offset = 0
	}

	for {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		if len(out) >= cfg.MaxRows {
			break
		}
		rows, err := queryPage(ctx, db, cfg, base, after, offset)
		if err != nil {
			return nil, nil, err
		}
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			if len(out) >= cfg.MaxRows {
				break
			}
			item, lastVal, err := c.rowToItem(cfg, row)
			if err != nil {
				logger.Warnf(ctx, "[sql] skip row: %v", err)
				continue
			}
			out = append(out, item)
			if lastVal != "" {
				next.LastValue = lastVal
			}
		}
		if after != "" {
			// incremental filter already applied; advance by page via last value
			after = next.LastValue
			offset = 0
			if len(rows) < cfg.PageSize {
				break
			}
		} else {
			offset += len(rows)
			if len(rows) < cfg.PageSize {
				break
			}
		}
	}
	return out, next, nil
}

func (c *Connector) rowToItem(cfg *Config, row map[string]any) (types.FetchedItem, string, error) {
	id := stringify(lookupColumn(row, cfg.IDColumn))
	if id == "" {
		return types.FetchedItem{}, "", fmt.Errorf("empty id column %q", cfg.IDColumn)
	}
	title := stringify(lookupColumn(row, cfg.TitleColumn))
	if title == "" {
		title = id
	}
	content := renderContent(cfg, row)
	updatedAt := time.Now().UTC()
	lastVal := ""
	if cfg.IncrementalColumn != "" {
		lastVal = stringify(lookupColumn(row, cfg.IncrementalColumn))
		if t, ok := parseFlexibleTime(lastVal); ok {
			updatedAt = t
		}
	}
	return types.FetchedItem{
		ExternalID:       "sql:" + id,
		Title:            title,
		Content:          []byte(content),
		ContentType:      "text/markdown",
		FileName:         sanitizeFileName(title) + ".md",
		UpdatedAt:        updatedAt,
		SourceResourceID: "sql:default",
		Metadata: map[string]string{
			"channel": types.ChannelSQL,
			"source":  "sql",
			"row_id":  id,
		},
	}, lastVal, nil
}

func renderContent(cfg *Config, row map[string]any) string {
	if tmpl := strings.TrimSpace(cfg.ContentTemplate); tmpl != "" {
		out := tmpl
		for k, v := range row {
			out = strings.ReplaceAll(out, "{{"+k+"}}", stringify(v))
		}
		return out
	}
	cols := cfg.ContentColumns
	if len(cols) == 0 {
		// default: all columns as markdown sections
		var b strings.Builder
		for k, v := range row {
			b.WriteString("## ")
			b.WriteString(k)
			b.WriteString("\n\n")
			b.WriteString(stringify(v))
			b.WriteString("\n\n")
		}
		return strings.TrimSpace(b.String())
	}
	var b strings.Builder
	for _, col := range cols {
		val := stringify(lookupColumn(row, col))
		if val == "" {
			continue
		}
		b.WriteString("## ")
		b.WriteString(col)
		b.WriteString("\n\n")
		b.WriteString(val)
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

func lookupColumn(row map[string]any, name string) any {
	if v, ok := row[name]; ok {
		return v
	}
	// case-insensitive fallback
	lower := strings.ToLower(name)
	for k, v := range row {
		if strings.ToLower(k) == lower {
			return v
		}
	}
	return nil
}

func stringify(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case []byte:
		return strings.TrimSpace(string(t))
	case time.Time:
		return t.UTC().Format(time.RFC3339)
	case fmt.Stringer:
		return t.String()
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprintf("%v", t)
		}
		s := string(b)
		if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
			return strings.Trim(s, `"`)
		}
		return s
	}
}

func parseFlexibleTime(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "row"
	}
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	name = replacer.Replace(name)
	if len(name) > 80 {
		name = name[:80]
	}
	return name
}

func decodeCursor(cursor *types.SyncCursor) *syncCursorState {
	if cursor == nil || cursor.ConnectorCursor == nil {
		return nil
	}
	b, err := json.Marshal(cursor.ConnectorCursor)
	if err != nil {
		return nil
	}
	var state syncCursorState
	if err := json.Unmarshal(b, &state); err != nil {
		return nil
	}
	return &state
}

func encodeCursor(state *syncCursorState) *types.SyncCursor {
	if state == nil {
		return nil
	}
	b, err := json.Marshal(state)
	if err != nil {
		return nil
	}
	m := map[string]interface{}{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	last := time.Now().UTC()
	if t, err := time.Parse(time.RFC3339, state.LastSyncTime); err == nil {
		last = t
	}
	return &types.SyncCursor{LastSyncTime: last, ConnectorCursor: m}
}
