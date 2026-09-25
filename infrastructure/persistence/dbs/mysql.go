package dbs

import (
	"naotodoserver/conf"
	"naotodoserver/infrastructure/logging"
	"naotodoserver/infrastructure/persistence/models"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitMySQL 初始化 MySQL 数据库连接
func InitMySQL() {
	// @step 1. 拼接数据库连接字符串 ("%s:%s@tcp(%s:%d)/%s")
	mysqlConfig := conf.Conf.MySQL
	openString := strings.Join([]string{
		mysqlConfig.Username, ":", mysqlConfig.Password,
		"@tcp(", mysqlConfig.Host, ":", mysqlConfig.Port, ")",
		"/", mysqlConfig.Database,
		"?charset=", mysqlConfig.Charset,
		"&parseTime=", mysqlConfig.ParseTime,
		"&loc=", mysqlConfig.Loc,
	}, "")

	// @step 2. 连接数据库
	db, err := gorm.Open(mysql.Open(openString), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(err)
	}

	// @step 3. 获取数据库连接池对象
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// @step 4. 设置链接属性
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// @step 5. 赋值数据库引用
	DB = db

	// @step 6. 创建数据库迁移 (测试)
	DoMigration()
}

// DoMigration 执行数据库迁移
func DoMigration() {
	DB.Set("gorm:table_options", "charset=utf8mb4")

	err := DB.AutoMigrate(
		models.User{},
		models.UserConfig{},
		models.UserSession{},
		models.Project{},
		models.ProjectPreference{},
		models.Tag{},
		models.TagPreference{},
		models.Task{},
		models.TaskCheckItem{},
		models.TaskComment{},
		models.Pomodoro{},
		models.PomodoroRecord{},
	)

	if err != nil {
		return
	}

	// 移除旧版「一用户一会话」唯一索引，支持多会话（幂等）
	var legacyUniqueCnt int64
	DB.Raw(
		"SELECT COUNT(*) FROM information_schema.statistics "+
			"WHERE table_schema = DATABASE() AND table_name = 'user_sessions' "+
			"AND index_name = 'idx_user_session_user_id' AND non_unique = 0",
	).Scan(&legacyUniqueCnt)
	if legacyUniqueCnt > 0 {
		if err := DB.Exec(
			"ALTER TABLE user_sessions DROP INDEX idx_user_session_user_id",
		).Error; err != nil {
			logging.Logger.Errorf(
				"移除 user_sessions 旧唯一索引失败，多设备登录将失效（请检查索引是否已被其他实例移除或权限不足）: %v",
				err,
			)
		}
	}

	// 增量同步复合索引 (user_id, updated_at, id)
	// 支持增量拉取的 WHERE user_id = ? AND (updated_at, id) > (?, ?) + ORDER BY updated_at ASC, id ASC；
	// 含 id 列使 keyset 同秒排序走索引，避免 filesort。
	// MySQL 索引名表内唯一即可，这里仍使用各表唯一命名，避免歧义。
	syncIndexTables := []struct {
		table string
		index string
	}{
		{"tasks", "idx_task_user_updated"},
		{"task_check_items", "idx_task_check_item_user_updated"},
		{"task_comments", "idx_task_comment_user_updated"},
		{"projects", "idx_project_user_updated"},
		{"tags", "idx_tag_user_updated"},
		{"pomodoros", "idx_pomodoro_user_updated"},
		{"pomodoro_records", "idx_pomodoro_record_user_updated"},
	}
	for _, item := range syncIndexTables {
		// 已存在则跳过（幂等）；CREATE INDEX 不支持 IF NOT EXISTS，用错误容忍
		var cnt int64
		DB.Raw(
			"SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
			item.table, item.index,
		).Scan(&cnt)
		if cnt > 0 {
			continue
		}
		if err := DB.Exec(
			"CREATE INDEX "+item.index+" ON "+item.table+" (user_id, updated_at, id)",
		).Error; err != nil {
			logging.Logger.
				WithFields(map[string]any{"index": item.index, "table": item.table}).
				Warnf("增量同步索引创建失败（增量拉取将退化为全表扫描）: %v", err)
			continue
		}
	}
}
