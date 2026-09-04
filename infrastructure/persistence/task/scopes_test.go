package task

import (
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"naotodoserver/infrastructure/persistence/models"
)

// fakeDialector 免连接构造 gorm.DB：scope 只是向 Statement 追加 clause，无需真 SQL 执行。
type fakeDialector struct{}

func (fakeDialector) Name() string                                        { return "fake" }
func (fakeDialector) Initialize(*gorm.DB) error                           { return nil }
func (fakeDialector) Migrator(*gorm.DB) gorm.Migrator                     { return nil }
func (fakeDialector) DataTypeOf(*schema.Field) string                     { return "text" }
func (fakeDialector) DefaultValueOf(*schema.Field) clause.Expression      { return nil }
func (fakeDialector) BindVarTo(w clause.Writer, _ *gorm.Statement, _ any) { w.WriteByte('?') }
func (fakeDialector) QuoteTo(w clause.Writer, s string) {
	w.WriteByte('`')
	w.WriteString(s)
	w.WriteByte('`')
}
func (fakeDialector) Explain(sql string, vars ...any) string { return sql }

// scopeDB 在 DryRun 会话上直接执行各 scope（scope 在真实流程中于 SQL 构建时依序调用，
// 此处直接调用以断言 clause 效果），返回 Statement 供 clause 级断言
func scopeDB(scopes ...func(*gorm.DB) *gorm.DB) (*gorm.Statement, error) {
	db, err := gorm.Open(fakeDialector{}, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	tx := db.Session(&gorm.Session{DryRun: true}).Model(&models.Task{})
	for _, scope := range scopes {
		tx = scope(tx)
	}
	return tx.Statement, nil
}

// whereExprs 取 WHERE clause 的表达式列表
func whereExprs(t *testing.T, stmt *gorm.Statement) []clause.Expression {
	t.Helper()
	w, ok := stmt.Clauses["WHERE"]
	if !ok {
		return nil
	}
	cond, ok := w.Expression.(clause.Where)
	if !ok || len(cond.Exprs) == 0 {
		return nil
	}
	return cond.Exprs
}

func TestByProjects(t *testing.T) {
	stmt, err := scopeDB(ByProjects(nil))
	if err != nil {
		t.Fatal(err)
	}
	if exprs := whereExprs(t, stmt); len(exprs) != 0 {
		t.Fatalf("空集应不过滤，实际 WHERE=%v", exprs)
	}

	for _, tc := range []struct{ ids []int64 }{{ids: []int64{11}}, {ids: []int64{11, 22, 33}}} {
		stmt, err := scopeDB(ByProjects(tc.ids))
		if err != nil {
			t.Fatal(err)
		}
		exprs := whereExprs(t, stmt)
		if len(exprs) != 1 {
			t.Fatalf("期望 1 条 WHERE，实际 %d", len(exprs))
		}
		expr, ok := exprs[0].(clause.Expr)
		if !ok || expr.SQL != "project_id IN ?" {
			t.Fatalf("project 过滤应为 IN，实际 %#v", exprs[0])
		}
		vars, ok := expr.Vars[0].([]int64)
		if !ok || len(vars) != len(tc.ids) {
			t.Fatalf("IN 参数应为 %d 个，实际 %#v", len(tc.ids), expr.Vars)
		}
	}
}

func TestByTags(t *testing.T) {
	stmt, err := scopeDB(ByTags(nil))
	if err != nil {
		t.Fatal(err)
	}
	if exprs := whereExprs(t, stmt); len(exprs) != 0 {
		t.Fatalf("空集应不过滤，实际 WHERE=%v", exprs)
	}

	stmt, err = scopeDB(ByTags([]string{"a", "b", "c"}))
	if err != nil {
		t.Fatal(err)
	}
	exprs := whereExprs(t, stmt)
	if len(exprs) != 1 {
		t.Fatalf("期望 1 条 WHERE，实际 %d", len(exprs))
	}
	expr, ok := exprs[0].(clause.Expr)
	if !ok || expr.SQL != "(tags LIKE ? OR tags LIKE ? OR tags LIKE ?)" {
		t.Fatalf("tag 多值应 OR 拼接，实际 %#v", exprs[0])
	}
	if len(expr.Vars) != 3 {
		t.Fatalf("OR 参数应为 3 个，实际 %d", len(expr.Vars))
	}
}

func TestProjectAndTagAreAnded(t *testing.T) {
	stmt, err := scopeDB(ByProjects([]int64{11}), ByTags([]string{"a"}))
	if err != nil {
		t.Fatal(err)
	}
	exprs := whereExprs(t, stmt)
	if len(exprs) != 2 {
		t.Fatalf("project 与 tag 应各占一条 WHERE（组间 AND），实际 %d: %#v", len(exprs), exprs)
	}
}

func TestByRelativeDateMonth(t *testing.T) {
	stmt, err := scopeDB(ByRelativeDate("month"))
	if err != nil {
		t.Fatal(err)
	}
	exprs := whereExprs(t, stmt)
	if len(exprs) != 1 {
		t.Fatalf("month 应产生 1 条 WHERE，实际 %d", len(exprs))
	}
	expr, ok := exprs[0].(clause.Expr)
	if !ok || !strings.Contains(expr.SQL, "end_at >= ?") || !strings.Contains(expr.SQL, "end_at < ?") {
		t.Fatalf("month 应为自然月闭窗口，实际 %#v", exprs[0])
	}
	if len(expr.Vars) != 2 {
		t.Fatalf("month 边界应为 2 个，实际 %d", len(expr.Vars))
	}
	start, ok1 := expr.Vars[0].(time.Time)
	end, ok2 := expr.Vars[1].(time.Time)
	if !ok1 || !ok2 {
		t.Fatalf("month 边界应为 time.Time，实际 %#v", expr.Vars)
	}
	now := time.Now()
	y, m, _ := now.Date()
	wantStart := time.Date(y, m, 1, 0, 0, 0, 0, time.Local)
	wantEnd := wantStart.AddDate(0, 1, 0)
	if !start.Equal(wantStart) || !end.Equal(wantEnd) {
		t.Fatalf("month 窗口 = [%v, %v)，期望 [%v, %v)", start, end, wantStart, wantEnd)
	}
}
