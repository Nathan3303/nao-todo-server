package dbs

import (
	"naotodoserver/conf"
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
	sqlDB.SetMaxOpenConns(100)
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
		models.Session{},
		models.Project{},
		models.ProjectPreference{},
		models.Tag{},
		models.TagPreference{},
		models.Task{},
		models.Event{},
		models.Comment{},
		models.CommentUser{},
	)

	if err != nil {
		return
	}
}
