package maimai

import (
	"bytes"
	"fmt"
	"os/exec"
	"tbot/utils/msg"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	e := zero.New()
	e.OnPrefix("b40").Handle(func(ctx *zero.Ctx) {
		args := ctx.State["args"].(string)
		var cmd *exec.Cmd
		if len(args) > 0 {
			cmd = exec.Command("python3", "lib3rd/maimai_b40/entry.py", "username", args)
		} else {
			cmd = exec.Command("python3", "lib3rd/maimai_b40/entry.py", "qq", fmt.Sprint(ctx.Event.Sender.ID))
		}
		bo := &bytes.Buffer{}
		be := &bytes.Buffer{}
		cmd.Stdout = bo
		cmd.Stderr = be
		err := cmd.Run()
		if err != nil {
			logrus.Error("spwan python error: ", err, "stderr: ", be.String())
			ctx.Send("执行出错，请稍后再试...")
			return
		}
		if len(bo.Bytes()) < 50 {
			ctx.Send("获取信息失败, 请确认已绑定用户并导入成绩:\nhttps://www.diving-fish.com/maimaidx/prober/")
			return
		}
		ctx.Send(msg.New().ImageBytes(bo.Bytes()))
	})

}
