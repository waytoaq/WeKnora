package sqlconn

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func openDB(cfg *Config) (*sql.DB, error) {
	dsn, driverName, err := buildDSN(cfg)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)
	timeout := time.Duration(cfg.QueryTimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func buildDSN(cfg *Config) (dsn string, driverName string, err error) {
	switch cfg.Driver {
	case "postgres":
		u := &url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(cfg.User, cfg.Password),
			Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Path:   "/" + cfg.Database,
		}
		q := url.Values{}
		ssl := cfg.SSLMode
		if ssl == "" {
			ssl = "disable"
		}
		q.Set("sslmode", ssl)
		u.RawQuery = q.Encode()
		return u.String(), "pgx", nil
	case "mysql":
		timeout := cfg.QueryTimeoutSec
		if timeout <= 0 {
			timeout = 60
		}
		userEsc := url.QueryEscape(cfg.User)
		passEsc := url.QueryEscape(cfg.Password)
		dbEsc := url.PathEscape(cfg.Database)
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&timeout=%ds&readTimeout=%ds&writeTimeout=%ds&charset=utf8mb4&loc=UTC",
			userEsc, passEsc, cfg.Host, cfg.Port, dbEsc, timeout, timeout, timeout,
		), "mysql", nil
	default:
		return "", "", fmt.Errorf("unsupported driver %q", cfg.Driver)
	}
}

func buildBaseQuery(cfg *Config) (string, error) {
	if cfg.Query != "" {
		return cfg.Query, nil
	}
	table := quoteIdent(cfg.Driver, cfg.Table)
	q := "SELECT * FROM " + table
	where := strings.TrimSpace(cfg.Where)
	if where != "" {
		lower := strings.ToLower(where)
		if strings.Contains(where, ";") || strings.Contains(lower, " drop ") || strings.Contains(lower, " delete ") {
			return "", fmt.Errorf("invalid where clause")
		}
		q += " WHERE " + where
	}
	return q, nil
}

func queryPage(
	ctx context.Context, db *sql.DB, cfg *Config, baseSQL string, afterValue string, offset int,
) ([]map[string]any, error) {
	timeout := time.Duration(cfg.QueryTimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	qctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	sqlText := baseSQL
	args := []any{}
	hasWhere := strings.Contains(strings.ToLower(baseSQL), " where ")

	if cfg.IncrementalColumn != "" && afterValue != "" {
		col := quoteIdent(cfg.Driver, cfg.IncrementalColumn)
		clause := fmt.Sprintf("%s > ?", col)
		if cfg.Driver == "postgres" {
			clause = fmt.Sprintf("%s > $%d", col, len(args)+1)
		}
		if hasWhere {
			sqlText += " AND " + clause
		} else {
			sqlText += " WHERE " + clause
		}
		args = append(args, afterValue)
		hasWhere = true
	}

	if cfg.IncrementalColumn != "" {
		col := quoteIdent(cfg.Driver, cfg.IncrementalColumn)
		sqlText += " ORDER BY " + col + " ASC"
	}

	limitClause := ""
	if cfg.Driver == "postgres" {
		limitClause = fmt.Sprintf(" LIMIT %d OFFSET %d", cfg.PageSize, offset)
	} else {
		limitClause = fmt.Sprintf(" LIMIT %d OFFSET %d", cfg.PageSize, offset)
	}
	sqlText += limitClause

	rows, err := db.QueryContext(qctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var out []map[string]any
	for rows.Next() {
		raw := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := map[string]any{}
		for i, col := range cols {
			row[col] = normalizeSQLValue(raw[i])
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func normalizeSQLValue(v any) any {
	switch t := v.(type) {
	case nil:
		return nil
	case []byte:
		return string(t)
	case time.Time:
		return t.UTC().Format(time.RFC3339)
	default:
		return t
	}
}
