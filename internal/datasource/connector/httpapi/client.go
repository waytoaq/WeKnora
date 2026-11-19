package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type client struct {
	httpClient *http.Client
	cfg        *Config
	headers    map[string]string
}

func newClient(cfg *Config) *client {
	headers := parseAuthHeaders(cfg.AuthHeaders)
	switch strings.ToLower(cfg.AuthType) {
	case "bearer":
		if tok := strings.TrimSpace(cfg.Token); tok != "" {
			headers["Authorization"] = "Bearer " + tok
		}
	case "header", "none", "":
		// custom headers only / none
	default:
		if tok := strings.TrimSpace(cfg.Token); tok != "" {
			headers["Authorization"] = "Bearer " + tok
		}
	}
	return &client{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		cfg:        cfg,
		headers:    headers,
	}
}

func parseAuthHeaders(raw string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:idx])
		v := strings.TrimSpace(line[idx+1:])
		if k != "" {
			out[k] = v
		}
	}
	return out
}

func (c *client) resolveURL(pathOrURL string) (string, error) {
	pathOrURL = strings.TrimSpace(pathOrURL)
	if pathOrURL == "" {
		return "", fmt.Errorf("empty url")
	}
	if strings.HasPrefix(pathOrURL, "http://") || strings.HasPrefix(pathOrURL, "https://") {
		return pathOrURL, nil
	}
	if !strings.HasPrefix(pathOrURL, "/") {
		pathOrURL = "/" + pathOrURL
	}
	return c.cfg.BaseURL + pathOrURL, nil
}

func (c *client) getJSON(ctx context.Context, rawURL string) (any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(body))
		if len(msg) > 300 {
			msg = msg[:300] + "…"
		}
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, msg)
	}
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}
	return payload, nil
}

func (c *client) listPage(ctx context.Context, page int) (any, error) {
	raw, err := c.resolveURL(c.cfg.ListPath)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if c.cfg.PageParam != "" {
		q.Set(c.cfg.PageParam, fmt.Sprintf("%d", page))
	}
	if c.cfg.PageSizeParam != "" && c.cfg.PageSize > 0 {
		q.Set(c.cfg.PageSizeParam, fmt.Sprintf("%d", c.cfg.PageSize))
	}
	u.RawQuery = q.Encode()
	return c.getJSON(ctx, u.String())
}

func (c *client) fetchDetail(ctx context.Context, id string) (any, error) {
	tmpl := strings.TrimSpace(c.cfg.DetailPathTemplate)
	if tmpl == "" {
		return nil, nil
	}
	path := strings.ReplaceAll(tmpl, "{id}", url.PathEscape(id))
	raw, err := c.resolveURL(path)
	if err != nil {
		return nil, err
	}
	return c.getJSON(ctx, raw)
}

func dig(payload any, path string) any {
	path = strings.TrimSpace(path)
	if path == "" {
		return payload
	}
	cur := payload
	for _, part := range strings.Split(path, ".") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur, ok = m[part]
		if !ok {
			return nil
		}
	}
	return cur
}

func asObjectList(v any) []map[string]any {
	switch t := v.(type) {
	case []any:
		out := make([]map[string]any, 0, len(t))
		for _, item := range t {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	case []map[string]any:
		return t
	default:
		return nil
	}
}

func fieldString(m map[string]any, key string) string {
	if key == "" || m == nil {
		return ""
	}
	raw, ok := m[key]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%v", v)
	case json.Number:
		return v.String()
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		s := string(b)
		if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
			return strings.Trim(s, `"`)
		}
		return s
	}
}
