package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"tbot/plugin/epicture"
	"tbot/utils"
	"tbot/utils/db"
)

func main() {
	cfg := utils.GetConfig()
	cfg.Init("config.yaml")
	db.InitDB()
	query := db.DB().Model(&epicture.Epicture{})
	log.Printf("Info total %v records", query.RowsAffected)
	rows, err := query.Rows()
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	for rows.Next() {
		var pic epicture.Epicture
		db.DB().ScanRows(rows, &pic)
		p, _ := filepath.Abs(utils.GetConfig().RuntimePath)
		path := fmt.Sprintf("%v/%v", p, pic.Path)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			log.Printf("ERR %v %v not exist", pic.ID, pic.Path)
			continue
		}
		if err != nil {
			log.Printf("ERR %v %v stat failed: %v", pic.ID, pic.Path, err)
			continue
		}
		if info.Size() == 0 {
			log.Printf("ERR %v %v size is 0", pic.ID, pic.Path)
			continue
		}
		log.Printf("OK %v %v size %v", pic.ID, pic.Path, info.Size())
	}
}
