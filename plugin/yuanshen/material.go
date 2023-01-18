package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const kBreak_Weapon = "1"
const kBreak_Cara = "2"

type Activity struct {
	BreakType     string   `json:"break_type"` // 1 武器 2 人物
	DropDay       []string `json:"drop_day"`
	Title         string
	ContentInfos  []map[string]interface{}
	ContentSource []map[string]interface{}
}
type ActivityList struct {
	Activities []Activity `json:"list"`
}
type ActivityResp struct {
	Data ActivityList `json:"data"`
	Msg  string       `json:"message"`
	Ret  int          `json:"retcode"`
}
type DropInfo struct {
	Cara map[string]bool
}

func init() {
	e := zero.New()

	e.OnCommand("今日op").Handle(func(ctx *zero.Ctx) {
		args := strings.Fields(ctx.State["args"].(string))

		print_cara := true
		print_weapon := true
		if len(args) > 0 {
			if args[0] == "天赋" {
				print_weapon = false
			} else if args[0] == "武器" {
				print_cara = false
			} else {
				ctx.Send("参数错误，使用方法：\n%今日op [天赋|武器]")
				return
			}
		}

		cli := &http.Client{}
		resp, err := cli.Get("https://api-static.mihoyo.com/common/blackboard/ys_obc/v1/get_activity_calendar?app_sn=ys_obc")
		if err != nil {
			ctx.Send("获取 API 信息失败")
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			ctx.Send("获取 API 信息失败, Read Error")
			return
		}
		var data ActivityResp
		err = json.Unmarshal(body, &data)
		if err != nil {
			ctx.Send("解析 API 返回失败")
			return
		}

		var result map[string]map[string]DropInfo = make(map[string]map[string]DropInfo)
		result[kBreak_Cara] = make(map[string]DropInfo)
		result[kBreak_Weapon] = make(map[string]DropInfo)

		// sunday: 0
		now := time.Now()
		var curday int = int(now.Weekday())
		if now.Hour() < 4 {
			curday = (curday + 7 - 1) % 7
		}
		if curday == 0 {
			curday = 7
		}
		logrus.Infof("今日OP curday %v", curday)
		for _, act := range data.Data.Activities {
			// logrus.Debugf("handling %+v", act)
			for _, day := range act.DropDay {
				if day == fmt.Sprint(curday) {
					title, ok := act.ContentInfos[0]["title"].(string)
					if !ok {
						ctx.Send("解析 json 失败")
						return
					}
					cata := strings.Split(title, "的")[0]
					if _, ok = result[act.BreakType][cata]; !ok {
						result[act.BreakType][cata] = DropInfo{
							Cara: make(map[string]bool),
						}
					}
					result[act.BreakType][cata].Cara[act.Title] = true
					break
				}
			}
		}

		PrintDrops := func(drops map[string]DropInfo) string {
			var sb strings.Builder
			for cata, dropinfo := range drops {
				sb.WriteString(fmt.Sprintf("======= %v ======\n", cata))
				keys := make([]string, 0)
				for key, _ := range dropinfo.Cara {
					keys = append(keys, key)
				}
				sb.WriteString(fmt.Sprintf("可以升: %v\n", strings.Join(keys, ", ")))
			}
			return sb.String()
		}
		var sb strings.Builder
		if print_cara {
			sb.WriteString("【今日天赋素材】\n")
			sb.WriteString(PrintDrops(result[kBreak_Cara]))
		}
		if print_weapon {
			sb.WriteString("【今日武器素材】\n")
			sb.WriteString(PrintDrops(result[kBreak_Weapon]))
		}
		logrus.Debugf(sb.String())
		ctx.Send(sb.String())
	})
}
