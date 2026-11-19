package httpapi

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

// Connector syncs records from a configurable HTTP list/detail JSON API.
type Connector struct{}

// NewConnector creates an HTTP API connector.
func NewConnector() *Connector { return &Connector{} }

// Type returns the connector type identifier.
func (c *Connector) Type() string { return types.ConnectorTypeHTTPAPI }

// Validate checks connectivity by fetching the first list page.
func (c *Connector) Validate(ctx context.Context, config *types.DataSourceConfig) error {
	cfg, err := parseConfig(config)
	if err != nil {
		return err
	}
	cli := newClient(cfg)
	payload, err := cli.listPage(ctx, 1)
	if err != nil {
		return fmt.Errorf("list endpoint: %w", err)
	}
	items := asObjectList(dig(payload, cfg.ItemsPath))
	if items == nil && cfg.ItemsPath != "" {
		return fmt.Errorf("%w: items_path %q did not resolve to an array", datasource.ErrInvalidConfig, cfg.ItemsPath)
	}
	return nil
}

// ListResources returns a single logical resource representing the configured API.
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
	name := cfg.ListPath
	if name == "" {
		name = cfg.BaseURL
	}
	return []types.Resource{{
		ExternalID:  "http_api:default",
		Name:        name,
		Type:        "api",
		Description: cfg.BaseURL,
		URL:         cfg.BaseURL,
	}}, nil
}

// ResolveResourceAncestors is a no-op for the flat HTTP API resource.
func (c *Connector) ResolveResourceAncestors(
	ctx context.Context, config *types.DataSourceConfig, resourceIDs []string,
) ([]string, error) {
	return []string{}, nil
}

// FetchAll downloads all pages of records.
func (c *Connector) FetchAll(
	ctx context.Context, config *types.DataSourceConfig, resourceIDs []string,
) ([]types.FetchedItem, error) {
	items, _, err := c.walk(ctx, config, nil, false)
	return items, err
}

// FetchIncremental returns records whose updated_at changed since the prior cursor.
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
	cli := newClient(cfg)
	next := &syncCursorState{
		LastSyncTime: time.Now().UTC().Format(time.RFC3339),
		ItemTimes:    map[string]string{},
	}
	var out []types.FetchedItem
	var pageErrors []string

	for page := 1; page <= cfg.MaxPages; page++ {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		payload, err := cli.listPage(ctx, page)
		if err != nil {
			pageErrors = append(pageErrors, fmt.Sprintf("page %d: %v", page, err))
			logger.Warnf(ctx, "[http_api] list page %d failed: %v", page, err)
			break
		}
		rows := asObjectList(dig(payload, cfg.ItemsPath))
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			if len(out) >= cfg.MaxItems {
				break
			}
			id := fieldString(row, cfg.IDField)
			if id == "" {
				continue
			}
			updated := fieldString(row, cfg.UpdatedAtField)
			next.ItemTimes[id] = updated
			if incremental && prev != nil && prev.ItemTimes != nil {
				if old, ok := prev.ItemTimes[id]; ok && old == updated && updated != "" {
					continue
				}
			}
			item, err := c.buildItem(ctx, cli, cfg, row, id)
			if err != nil {
				logger.Warnf(ctx, "[http_api] build item %s failed: %v", id, err)
				continue
			}
			out = append(out, item)
		}
		if len(out) >= cfg.MaxItems || len(rows) < cfg.PageSize {
			break
		}
	}

	if incremental && prev != nil {
		for id, ts := range prev.ItemTimes {
			if _, ok := next.ItemTimes[id]; !ok {
				next.ItemTimes[id] = ts
			}
		}
	}

	if len(out) == 0 && len(pageErrors) > 0 {
		return nil, nil, fmt.Errorf("http_api sync failed: %s", strings.Join(pageErrors, "; "))
	}
	return out, next, nil
}

func (c *Connector) buildItem(
	ctx context.Context, cli *client, cfg *Config, row map[string]any, id string,
) (types.FetchedItem, error) {
	src := row
	if cfg.DetailPathTemplate != "" {
		detail, err := cli.fetchDetail(ctx, id)
		if err != nil {
			return types.FetchedItem{}, err
		}
		if m, ok := detail.(map[string]any); ok {
			src = m
		} else if m := asSingleObject(dig(detail, cfg.ItemsPath)); m != nil {
			src = m
		}
	}
	title := fieldString(src, cfg.TitleField)
	if title == "" {
		title = id
	}
	content := fieldString(src, cfg.ContentField)
	if content == "" {
		b, _ := json.MarshalIndent(src, "", "  ")
		content = "```json\n" + string(b) + "\n```"
	}
	updatedAt := parseTime(fieldString(src, cfg.UpdatedAtField))
	itemURL := fieldString(src, cfg.URLField)
	return types.FetchedItem{
		ExternalID:       "http_api:" + id,
		Title:            title,
		Content:          []byte(content),
		ContentType:      "text/markdown",
		FileName:         sanitizeFileName(title) + ".md",
		URL:              itemURL,
		UpdatedAt:        updatedAt,
		SourceResourceID: "http_api:default",
		Metadata: map[string]string{
			"channel": types.ChannelHTTPAPI,
			"source":  "http_api",
			"item_id": id,
		},
	}, nil
}

func asSingleObject(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	list := asObjectList(v)
	if len(list) == 1 {
		return list[0]
	}
	return nil
}

func parseTime(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Now().UTC()
	}
	layouts := []string{
		time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC()
		}
	}
	return time.Now().UTC()
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "item"
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
	if state.ItemTimes == nil {
		state.ItemTimes = map[string]string{}
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
