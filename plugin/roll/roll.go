package cmd

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	e := zero.New()
	rand.Seed(time.Now().UnixNano())
	e.OnCommand("roll").Handle(func(ctx *zero.Ctx) {
		var err error
		args := strings.Fields(ctx.State["args"].(string))
		defer func() {
			if err != nil {
				ctx.Send(err.Error())
			}
		}()
		if len(args) != 1 {
			err = fmt.Errorf("参数错误, 使用例：\n%%roll 6\n%%roll 3d6")
			return
		}
		v := args[0]
		n, err := strconv.Atoi(v)
		if err == nil {
			if n <= 0 {
				err = fmt.Errorf("?")
			} else {
				ctx.Send(rand.Intn(int(n-1)) + 1)
			}
			return
		}
		d := strings.Split(args[0], "d")
		if len(d) != 2 {
			err = fmt.Errorf("?")
			return
		}
		cnt, err := strconv.Atoi(d[0])
		if err != nil || cnt <= 0 {
			err = fmt.Errorf("?")
			return
		}
		n, err = strconv.Atoi(d[1])
		if err != nil || n <= 0 {
			err = fmt.Errorf("?")
			return
		}
		result := make([]string, 0, cnt)
		for i := 0; i < cnt; i++ {
			result = append(result, fmt.Sprint(rand.Intn(n-1)+1))
		}
		ctx.Send(strings.Join(result, " "))
	})

	e.OnCommand("rand").Handle(func(ctx *zero.Ctx) {
		args := strings.Fields(ctx.State["args"].(string))
		if len(args) == 0 {
			ctx.Send("没东西可 rand")
			return
		}
		if len(args) == 1 {
			ctx.Send("就一个还 rand？")
			return
		}
		ctx.Send(args[rand.Intn(len(args))])
	})
}
