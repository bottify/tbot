package epicture

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"tbot/utils"
	"tbot/utils/db"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"gorm.io/gorm"
)

var once sync.Once

func init() {
	e := zero.New()
	e.OnCommand("涩图").Handle(func(ctx *zero.Ctx) {
		once.Do(Setup)
	})

	e.OnCommand("上传涩图").Handle(func(ctx *zero.Ctx) {
		once.Do(Setup)
		cnt, succ := 0, 0
		for _, msg := range ctx.Event.Message {
			if msg.Type == "image" {
				file, ok := msg.Data["file"]
				if ok {
					if err := SaveOneImageByFile(ctx, file); err == nil {
						succ++
					} else {
						log.Errorf("SaveOneImageByFile failed, file [%v] err %v", file, err)
					}
				}
				cnt++
			}
		}
		if cnt == 0 {
			ctx.Send("? 图来")
		} else {
			ctx.Send(fmt.Sprintf("一共识别出 %v 张涩图，成功上传 %v 张", cnt, succ))
		}
	})
}

type Epicture struct {
	gorm.Model
	Path       string
	Category   string `gorm:"index;default:default"`
	Comment    string
	Status     int   `gorm:"index"`
	UploadFrom int64 `gorm:"index;default:0"`
	UploaderID int64 `gorm:"index"`
}

func Setup() {
	db := db.DB()
	err := db.AutoMigrate(&Epicture{})
	if err != nil {
		log.Error("db AutoMigrate failed: ", err)
	} else {
		log.Debug("db AutoMigrate &Epicture{} succeed.")
	}
	err = os.MkdirAll(fmt.Sprintf("%v/data/tbot", utils.GetConfig().RuntimePath), os.FileMode(0770))
	if err != nil {
		log.Error("mkdir: ", err)
	}
}

func genFileIdForSave(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx >= 0 && idx < len(path)-1 {
		path = path[idx+1:]
	}
	segs := strings.Split(path, ".")
	if len(segs) == 1 {
		return segs[0]
	}
	return fmt.Sprintf("%v.%v", segs[0], segs[len(segs)-1])
}
func runtimePath(relative string) string {
	return fmt.Sprintf("%v/%v", utils.GetConfig().RuntimePath, relative)
}

func SaveOneImageByFile(ctx *zero.Ctx, fileid string) error {
	result := ctx.GetImage(fileid)
	path := result.Get("file").String()
	if len(path) > 0 {
		new_path := fmt.Sprintf("data/tbot/%v", genFileIdForSave(path))
		err := os.Rename(runtimePath(path), runtimePath(new_path))
		if err != nil {
			log.Errorf("move image [%v] from [%v] -> [%v], err %v", fileid, path, new_path, err)
		} else {
			log.Debugf("move image [%v] from [%v] -> [%v], err %v", fileid, path, new_path, err)
			err = db.DB().Create(&Epicture{
				Path:       new_path,
				UploadFrom: ctx.Event.GroupID,
				UploaderID: ctx.Event.Sender.ID,
			}).Error
		}
		return err
	}
	return fmt.Errorf("get image file failed, empty path")
}
