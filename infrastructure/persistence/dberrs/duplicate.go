// Package dberrs 提供数据库驱动错误的统一判定（SEC-12）。
package dberrs

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

// IsDuplicateKey 判断是否为唯一索引 / 主键冲突（MySQL 1062）。
// 迁移自 identity 用户仓储的同名私有函数，判定范围保持不变。
// @param err 数据库返回的错误
// @return 命中唯一约束冲突返回 true
func IsDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
