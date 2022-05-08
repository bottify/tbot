package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tbot/plugin/epicture"
	"tbot/utils"
	"tbot/utils/db"
)

var fileList = flag.String("list", "", "file list")
var group = flag.Int64("group", 0, "group")
var uploader = flag.Int64("uploader", 0, "uploader")
var config = flag.String("config", "config.yaml", "config")

func FileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return "", err
	}
	hash := md5.New()
	_, _ = io.Copy(hash, file)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func main() {
	flag.Parse()
	if *fileList == "" {
		fmt.Println("Usage: import_img -list <file list> -group <group> -uploader <uploader>")
		os.Exit(0)
	}

	cfg := utils.GetConfig()
	err := cfg.Init(*config)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	db.InitDB()

	file, err := os.OpenFile(*fileList, os.O_RDONLY, 0664)
	if err != nil {
		log.Fatalf("ERR Open file %v failed: %v", *fileList, err)
		os.Exit(1)
	}
	log.Printf("INFO start import img for group %v uploader %v", *group, *uploader)
	log.Printf("INFO img will be import to data path: %v", cfg.GetDataPath(""))
	// wait for confirm
	fmt.Print("Press Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	buf := bufio.NewReader(file)
	for {
		line, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("ERR read file %v failed: %v", *fileList, err)
			os.Exit(1)
		}
		img := strings.TrimSpace(line)
		md5str, err := FileMD5(img)
		if err != nil {
			log.Printf("ERR %v calc md5 failed", img)
			continue
		}
		suffix := filepath.Ext(img)
		err = os.Rename(img, cfg.GetDataPath(fmt.Sprintf("%v.%v", md5str, suffix)))
		if err != nil {
			log.Printf("ERR %v move to datadir failed: %v", img, err)
			continue
		}
		item := &epicture.Epicture{
			Path:       fmt.Sprintf("data/tbot/%v%v", md5str, suffix),
			UploadFrom: *group,
			UploaderID: *uploader,
		}
		err = db.DB().Create(item).Error
		if err != nil {
			log.Printf("ERR %v insert db failed: %v", img, err)
			continue
		}
		log.Printf("OK id %v %v -> %v", item.ID, img, fmt.Sprintf("data/tbot/%v%v", md5str, suffix))
	}
}
