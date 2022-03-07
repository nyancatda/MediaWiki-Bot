package SQLDB

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

/**
 * @description: 创建全局数据库链接
 * @param {*}
 * @return {*gorm.DB}
 * @return {error}
 */
func SQLDBLink() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("data/database.db"), &gorm.Config{})
	if err != nil {
		log.Println(err)
	}
	db.AutoMigrate(&UserInfo{})
	DB = db
	return db, err
}
