package main

import (
	"flag"
	"tbot/utils"
	"tbot/utils/db"
)

var fileList = flag.String("list", "", "file list")

func main() {
	cfg := utils.GetConfig()
	cfg.Init("config.yaml")
	db.InitDB()

	// TODO

}
