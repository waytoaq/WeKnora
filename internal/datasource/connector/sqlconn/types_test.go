package sqlconn

import (
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
)

func TestEnsureReadOnlySelect(t *testing.T) {
	if err := ensureReadOnlySelect("SELECT * FROM t"); err != nil {
		t.Fatalf("select should pass: %v", err)
	}
	if err := ensureReadOnlySelect("WITH x AS (SELECT 1) SELECT * FROM x"); err != nil {
		t.Fatalf("cte should pass: %v", err)
	}
	if err := ensureReadOnlySelect("DELETE FROM t"); err == nil {
		t.Fatal("delete must fail")
	}
	if err := ensureReadOnlySelect("SELECT 1; DROP TABLE t"); err == nil {
		t.Fatal("multi-statement must fail")
	}
}

func TestParseConfigRequiresQueryOrTable(t *testing.T) {
	_, err := parseConfig(&types.DataSourceConfig{
		Credentials: map[string]interface{}{
			"driver": "postgres", "host": "h", "database": "d", "user": "u",
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
