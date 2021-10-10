package epicture

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"tbot/utils"
	"tbot/utils/db"
	"tbot/utils/msg"
	"tbot/utils/rules"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

var once sync.Once

func init() {
	e := zero.New()
	e.OnCommand("涩图").Handle(func(ctx *zero.Ctx) {
		once.Do(Setup)
		var pic Epicture
		query := db.DB().Order("random()").Where("upload_from = ?", ctx.Event.GroupID)
		if ctx.Event.DetailType == "private" {
			query = query.Where("uploader_id = ?", ctx.Event.Sender.ID)
		}
		err := query.First(&pic).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Send("一张涩图都还没有哦...")
			return
		} else if err != nil {
			ctx.Send("执行错误，请联系管理员查看日志")
			return
		}
		p, _ := filepath.Abs(utils.GetConfig().RuntimePath)
		path := fmt.Sprintf("file://%v/%v", p, pic.Path)
		ctx.Send(msg.New().Text(fmt.Sprintf("id: %v\n%v", pic.ID, pic.Comment)).Image(path))
	})

	e.OnCommand("涩图存量").Handle(func(ctx *zero.Ctx) {
		once.Do(Setup)
		r := db.DB().Model(&Epicture{}).Select("upload_from, category, count(1) as cnt").Group("category").Having("upload_from = ?", ctx.Event.GroupID)
		rows, err := r.Rows()
		defer rows.Close()
		if err != nil {
			log.Errorf("query count for group [%v] error: %v", ctx.Event.GroupID, err)
			ctx.Send("查询出错，请联系管理员查看日志")
		} else {
			sb := &bytes.Buffer{}
			for rows.Next() {
				var from, category string
				var cnt int
				err = rows.Scan(&from, &category, &cnt)
				if err != nil {
					log.Errorf("scan row error: %v", err)
					continue
				}
				log.Debugf("scaned: [%v %v %v]", from, category, cnt)
				sb.WriteString("\n")
				sb.WriteString(fmt.Sprintf("%v: %v 张", category, cnt))
			}
			if sb.Len() == 0 {
				ctx.Send("目前还没有任何一张涩图...")
				return
			}
			prefix := ""
			if ctx.Event.DetailType == "group" {
				prefix = "你群的"
			} else if ctx.Event.DetailType == "private" {
				prefix = "你的"
			}
			ctx.Send(fmt.Sprintf("%v涩图存量：%v", prefix, sb.String()))
		}
	})

	e.OnMessage(rules.CommandWithReply("上传涩图")).Handle(func(ctx *zero.Ctx) {
		for _, msg := range ctx.Event.Message {
			if msg.Type == "reply" {
				id, _ := strconv.Atoi(msg.Data["id"])
				actual_msg := ctx.GetMessage(int64(id))
				if actual_msg.Elements != nil && len(actual_msg.Elements) > 0 {
					cnt, succ := 0, 0
					if actual_msg.Elements[0].Type != "forward" {
						log.Debug("parsed replyed msg of: ", actual_msg.Elements)
						cnt, succ = SaveImageInMessage(ctx, actual_msg.Elements)
					} else {
						msgs := message.Message{}
						data := ctx.GetForwardMessage(actual_msg.Elements[0].Data["id"])
						data.Get("messages").ForEach(func(_, value gjson.Result) bool {
							msgs = append(msgs, message.ParseMessageFromArray(value.Get("content"))...)
							return true
						})
						log.Debug("parsed replyed forwarded msg id, parsed msg len %v", len(msgs))
						cnt, succ = SaveImageInMessage(ctx, msgs)
					}
					if cnt == 0 {
						ctx.Send("？没识别到任何一张图")
					} else {
						ctx.Send(fmt.Sprintf("识别到 %v 张图片，上传成功 %v 张", cnt, succ))
					}
					return
				} else {
					log.Errorf("get reply msg failed! src [%v] id [%v] target [%+v]", id, msg, id, actual_msg)
				}
			} else {
				continue
			}
		}
		log.Error("logic error, unexpected handler call with msg: ", ctx.Event.RawMessage)
		ctx.Send("发生内部逻辑错误")
	})

	e.OnCommand("上传涩图").Handle(func(ctx *zero.Ctx) {
		once.Do(Setup)
		cnt, succ := SaveImageInMessage(ctx, ctx.Event.Message)
		if cnt == 0 {
			ctx.Send("？没识别到任何一张图")
		} else {
			ctx.Send(fmt.Sprintf("识别到 %v 张图片，上传成功 %v 张", cnt, succ))
		}
	})
}

func SaveImageInMessage(ctx *zero.Ctx, msg message.Message) (cnt int, succ int) {
	log.Debug("handle real msg segment len ", len(ctx.Event.Message))
	cnt, succ = 0, 0
	for _, msg := range msg {
		log.Debug("segment: ", msg.Type, msg.Data)
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
	log.Info("SaveImageInMessage msg cnt %v img %v succ %v", len(msg), cnt, succ)
	return cnt, succ
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
