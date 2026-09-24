package project

import (
	"testing"

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

// scopeDB 在 DryRun 会话上直接执行各 scope，返回 Statement 供 clause 级断言
func scopeDB(scopes ...func(*gorm.DB) *gorm.DB) (*gorm.Statement, error) {
	db, err := gorm.Open(fakeDialector{}, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	tx := db.Session(&gorm.Session{DryRun: true}).Model(&models.Project{})
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

// TestByProjectArchived 归档谓词回归（DP-1=(b)）：
// true ⇒ archived_at IS NOT NULL；false ⇒ archived_at IS NULL（修复前清单列表无归档过滤）。
func TestByProjectArchived(t *testing.T) {
	for _, tc := range []struct {
		name     string
		archived bool
		wantSQL  string
	}{
		{name: "true 仅已归档", archived: true, wantSQL: "archived_at IS NOT NULL"},
		{name: "显式 false 排除归档", archived: false, wantSQL: "archived_at IS NULL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stmt, err := scopeDB(ByProjectArchived(tc.archived))
			if err != nil {
				t.Fatal(err)
			}
			exprs := whereExprs(t, stmt)
			if len(exprs) != 1 {
				t.Fatalf("期望 1 条 WHERE，实际 %d", len(exprs))
			}
			expr, ok := exprs[0].(clause.Expr)
			if !ok || expr.SQL != tc.wantSQL {
				t.Fatalf("期望 %q，实际 %#v", tc.wantSQL, exprs[0])
			}
		})
	}
}

// TestByProjectArchivedUnsetIsZeroValue 固化「未传 isArchived」与显式 false 同路径：
// bool 值类型无三态 ⇒ 未传即零值 false，两者产生同一谓词（不存在隐式分叉）。
func TestByProjectArchivedUnsetIsZeroValue(t *testing.T) {
	var unset bool // 模拟「未传」
	unsetStmt, err := scopeDB(ByProjectArchived(unset))
	if err != nil {
		t.Fatal(err)
	}
	explicitStmt, err := scopeDB(ByProjectArchived(false))
	if err != nil {
		t.Fatal(err)
	}
	unsetExprs := whereExprs(t, unsetStmt)
	explicitExprs := whereExprs(t, explicitStmt)
	if len(unsetExprs) != 1 || len(explicitExprs) != 1 {
		t.Fatalf("应各 1 条 WHERE，实际 unset=%d explicit=%d", len(unsetExprs), len(explicitExprs))
	}
	unsetExpr, ok1 := unsetExprs[0].(clause.Expr)
	explicitExpr, ok2 := explicitExprs[0].(clause.Expr)
	if !ok1 || !ok2 {
		t.Fatalf("谓词类型错误: unset=%#v explicit=%#v", unsetExpr, explicitExpr)
	}
	if unsetExpr.SQL != "archived_at IS NULL" || unsetExpr.SQL != explicitExpr.SQL {
		t.Fatalf("未传与显式 false 应同谓词 IS NULL，实际 unset=%q explicit=%q", unsetExpr.SQL, explicitExpr.SQL)
	}
}
