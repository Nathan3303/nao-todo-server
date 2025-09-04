package core

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectDB(user string, password string, host string, port int, dbname string) {
	var openString = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4", user, password, host, port, dbname)
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

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
}
