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
	"sync/atomic"
	"tbot/utils"
	"tbot/utils/db"
	"tbot/utils/msg"
	"tbot/utils/rules"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

var once sync.Once
var epicCounter sync.Map

// var groupQuota map[int64]int64

func init() {
	e := zero.New()
	// groupQuota = make(map[int64]int64)
	// groupQuota[328992326] = 15

	go func() {
		last_hr := -1
		for {
			hr := time.Now().Hour()
			if hr != last_hr {
				epicCounter.Range(func(k, v interface{}) bool {
					log.Info("clearing epic counter for group: ", k)
					atomic.AddInt64(v.(*int64), -1*atomic.LoadInt64(v.(*int64)))
					return true
				})
				last_hr = hr
			}
			time.Sleep(time.Second)
		}
	}()

	e.OnCommand("涩图帮助").Handle(func(ctx *zero.Ctx) {
		once.Do(Setup)
		var sb strings.Builder
		sb.WriteString("涩图 - epic(ture) v0.3  // tag, comment 功能绝赞开发中...\n")
		sb.WriteString("-------------- 目前指令 ---------------\n")
		sb.WriteString("%涩图        来一份你群的随机涩图\n")
		sb.WriteString("%涩图存量    看看库存\n")
		sb.WriteString("%上传涩图    上传同一条消息内的图片(移动端可在发送图片时上滑弹出文本输入)，也可以通过[引用]别人发的图片消息或者转发消息来上传\n")
		ctx.Send(sb.String())
	})

	e.OnCommand("涩图").Handle(func(ctx *zero.Ctx) {
		once.Do(Setup)
		var pic Epicture
		var query *gorm.DB
		var extinfo string

		v, _ := epicCounter.LoadOrStore(ctx.Event.GroupID, new(int64))
		cnt := atomic.AddInt64(v.(*int64), 1)
		// quota, ok := groupQuota[ctx.Event.GroupID]
		quota, ok := utils.GetConfig().GetGroupEpicQuota(ctx.Event.GroupID)
		if ok && cnt > quota {
			ctx.Send("提示: 不可以色色（你群的涩图额度已经用完，每小时重置）")
			return
		}

		if utils.Contains(utils.GetConfig().GetSuperUsers(), fmt.Sprint(ctx.Event.Sender.ID)) {
			arg, _ := ctx.State["args"].(string)
			id, err := strconv.Atoi(arg)
			log.Debug(arg, ";", id, ";", err)
			if len(arg) > 0 && err == nil {
				query = db.DB().Where("id = ?", id)
				extinfo = "[超管特权]"
				log.Debug("SuperUser mode on, specify hpic id: ", id)
			}
		}
		if query == nil {
			query = db.DB().Order("random()").Where("upload_from = ?", ctx.Event.GroupID)
			if ctx.Event.DetailType == "private" {
				query = query.Where("uploader_id = ?", ctx.Event.Sender.ID)
			}
		}
		err := query.First(&pic).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if ctx.Event.DetailType == "group" {
				ctx.Send("你群现在还一张涩图都还没有，不如先用上传涩图来点?")
			} else {
				ctx.Send("你现在还一张涩图都还没有，不如先用上传涩图来点?")
			}
			return
		} else if err != nil {
			ctx.Send("执行错误，请联系管理员查看日志")
			return
		}
		idx := strings.LastIndex(pic.Path, "/")
		link := fmt.Sprintf("http://jk.tamce.cn/img/%v", pic.Path[idx+1:idx+17])

		p, _ := filepath.Abs(utils.GetConfig().RuntimePath)
		path := fmt.Sprintf("file://%v/%v", p, pic.Path)
		ctx.Send(msg.New().Text(fmt.Sprintf("%v[%v]%v\n%v", extinfo, pic.ID, link, pic.Comment)).Image(path))
		// qr, err := qrcode.Encode(link, qrcode.Low, 256)
		// ctx.Send(msg.New().Text(fmt.Sprintf("%vid: %v\n%v\n%v", extinfo, pic.ID, link, pic.Comment)).ImageBytes(qr))
	})

	e.OnCommand("涩图存量").Handle(func(ctx *zero.Ctx) {
		once.Do(Setup)
		r := db.DB().Model(&Epicture{}).Select("upload_from, category, count(1) as cnt").Group("upload_from, category").Having("upload_from = ?", ctx.Event.GroupID)
		log.Debug(r.Statement.SQL.String())
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
			prefix := ""
			if ctx.Event.DetailType == "group" {
				prefix = "你群的"
			} else if ctx.Event.DetailType == "private" {
				prefix = "你的"
			}
			if sb.Len() == 0 {
				ctx.Send(fmt.Sprintf("%v目前还没有任何一张涩图...", prefix))
				return
			}
			ctx.Send(fmt.Sprintf("%v涩图存量：%v", prefix, sb.String()))
		}
	})

	e.OnMessage(rules.CommandWithReply("上传涩图")).Handle(func(ctx *zero.Ctx) {
		once.Do(Setup)
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
						extra := ""
						if cnt != succ {
							extra = "，可能是有图被夹了或者下载出问题了"
						}
						ctx.Send(fmt.Sprintf("识别到 %v 张图片，上传成功 %v 张%v", cnt, succ, extra))
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
			extra := ""
			if cnt != succ {
				extra = "，可能是有图被夹了或者下载出问题了"
			}
			ctx.Send(fmt.Sprintf("识别到 %v 张图片，上传成功 %v 张%v", cnt, succ, extra))
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
		if stat, err := os.Stat(runtimePath(path)); err != nil {
			log.Errorf("fileid [%v] file [%v] stat error: %v", fileid, path, err)
			return fmt.Errorf("fileid [%v] file [%v] stat error: %v", fileid, path, err)
		} else {
			if stat.Size() == 0 {
				log.Errorf("fileid [%v] file [%v] is empty", fileid, path)
				return fmt.Errorf("get image file [%v][%v] failed, file size 0", fileid, path)
			}
		}

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
