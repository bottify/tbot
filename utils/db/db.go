package db

import (
	"fmt"
	"tbot/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB() {
	cfg := utils.GetConfig()
	d, err := gorm.Open(sqlite.Open(cfg.DBFile), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprint("open db failed: ", err))
	}
	db = d
}

func DB() *gorm.DB {
	return db
}
