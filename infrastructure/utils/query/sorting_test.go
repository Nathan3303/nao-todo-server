package query

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// sortDryRunSQL 用 mysql DryRun 生成带 ORDER BY 的 SQL（不连接数据库）
func sortDryRunSQL(t *testing.T, sort string) string {
	t.Helper()
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "user:pass@tcp(127.0.0.1:3306)/db?parseTime=true",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	type stub struct {
		ID        int64
		Name      string
		SortId    uint16
		CreatedAt time.Time
	}
	stmt := db.Model(&stub{}).Scopes(Sort(sort)).Find(&stub{}).Statement
	return db.Explain(stmt.SQL.String(), stmt.Vars...)
}

// assertNoOrderBy 非法 field / direction 必须回落到「无 ORDER BY」，不得拼接进 SQL
func assertNoOrderBy(t *testing.T, sort string) {
	t.Helper()
	sql := sortDryRunSQL(t, sort)
	if strings.Contains(strings.ToLower(sql), "order by") {
		t.Fatalf("sort=%q 应被拒绝/回落（无 ORDER BY），实际 SQL: %s", sort, sql)
	}
}

func TestSortWhitelistAllowsKnownFields(t *testing.T) {
	for _, tc := range []struct {
		sort string
		want string
	}{
		{sort: "name:asc", want: "order by name asc"},
		{sort: "sortId:desc", want: "order by sort_id desc"},
		{sort: "createdAt:desc", want: "order by created_at desc"},
		{sort: "startAt:asc", want: "order by start_at asc"},
		{sort: "created_at:asc", want: "order by created_at asc"}, // 兼容 snake_case 写法
	} {
		t.Run(tc.sort, func(t *testing.T) {
			sql := strings.ToLower(sortDryRunSQL(t, tc.sort))
			if !strings.Contains(sql, tc.want) {
				t.Fatalf("sort=%q 期望含 %q，实际 SQL: %s", tc.sort, tc.want, sql)
			}
		})
	}
}

func TestSortWhitelistNormalizesDirectionCase(t *testing.T) {
	for _, sort := range []string{"name:ASC", "name:Desc"} {
		t.Run(sort, func(t *testing.T) {
			sql := strings.ToLower(sortDryRunSQL(t, sort))
			if !strings.Contains(sql, "order by name ") {
				t.Fatalf("sort=%q 方向应归一为小写，实际 SQL: %s", sort, sql)
			}
			if strings.Contains(sql, "asc;") || strings.Contains(sql, ":desc") {
				t.Fatalf("sort=%q 方向不应原样拼接，实际 SQL: %s", sort, sql)
			}
		})
	}
}

func TestSortWhitelistRejectsInjectionAndIllegalInput(t *testing.T) {
	for _, sort := range []string{
		"",                          // 空值 = 不过滤
		"id;SELECT 1:asc",           // SQL 注入样例（子项② 实测样本）
		"name:asc;DROP TABLE tasks", // 方向注入
		"name:asc,id:desc",          // 多语句/逗号
		"name:asc:desc",             // 段数非法
		"name:",                     // 缺方向
		":asc",                      // 缺字段
		"unknownField:asc",          // 白名单外字段
		"name:sideways",             // 白名单外方向
		"users.password:asc",        // 跨表字段
	} {
		t.Run(sort, func(t *testing.T) {
			assertNoOrderBy(t, sort)
		})
	}
}
