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
	if len(os.Args) < 2 {
		fmt.Println("Usage: check_db <config>")
		os.Exit(0)
	}

	cfg := utils.GetConfig()
	err := cfg.Init(os.Args[1])
	if err != nil {
		log.Fatalf("error init config [%v]: %v", os.Args[1], err)
		os.Exit(1)
	}
	db.InitDB()
	query := db.DB().Model(&epicture.Epicture{})
	rows, err := query.Rows()
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	var groupPics map[int64](map[string]uint) = make(map[int64](map[string]uint))
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
		if _, ok := groupPics[pic.UploadFrom]; !ok {
			groupPics[pic.UploadFrom] = make(map[string]uint)
		}
		pics := groupPics[pic.UploadFrom]
		if existed_pic, ok := pics[pic.Path]; ok {
			log.Printf("WARN %v %v existed for group %v previous %v", pic.ID, pic.Path, pic.UploadFrom, existed_pic)
		} else {
			pics[pic.Path] = pic.ID
		}
		log.Printf("OK %v %v group %v size %v", pic.ID, pic.Path, pic.UploadFrom, info.Size())
	}
}
