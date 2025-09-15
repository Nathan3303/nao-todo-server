package core

import (
	"fmt"
	"naotodoserver/flags"
	"naotodoserver/globals"
	"naotodoserver/models"
	"os"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB(user string, password string, host string, port int, dbname string) {
	var openString = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbname)
	db, err := gorm.Open(mysql.Open(openString), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		logrus.Error("数据库连接失败")
	} else {
		logrus.Info("数据库连接成功")
	}

	sqlDB, err := db.DB()
	if err != nil {
		logrus.Error("db.DB() Error:", err)
	}

	// 设置链接属性
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 初始化雪花 ID 生成器
	globals.Vars.SnowNode = initSnowflake(1)

	DB = db
	globals.Vars.DB = db
}

func initSnowflake(machineID int64) *snowflake.Node {
	// 设置纪元时间（可选，避免2039问题）
	snowflake.Epoch = time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC).Unix() * 1000

	var err error
	SnowNode, err := snowflake.NewNode(machineID)
	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}
	return SnowNode
}

func CheckAndExecuteAutoMigration() {
	if !flags.Flags.AutoMigration {
		return
	}

	err := DB.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.UserConfig{},
		&models.Project{},
		&models.ProjectPreference{},
		&models.Tag{},
		&models.TagPreference{},
		&models.Todo{},
		&models.Event{},
		&models.Comment{},
	)
	if err != nil {
		logrus.Fatal("数据库自动迁移失败")
	}

	os.Exit(0)
}
