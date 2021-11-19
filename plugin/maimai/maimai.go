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

	e.OnCommand("底分分析").Handle(func(ctx *zero.Ctx) {
		args := ctx.State["args"].(string)
		var result *MaimaiAnalysisResult
		if len(args) > 0 {
			result = GetMinMaxChart(args, "")
		} else {
			qq := fmt.Sprint(ctx.Event.Sender.ID)
			result = GetMinMaxChart("", qq)
		}
		if result == nil {
			ctx.Send("分析失败，请稍后再重试下，或者确认下已经绑定了 qq: https://www.diving-fish.com/maimaidx/prober/")
			return
		}
		ctx.Send(fmt.Sprintf(
			`==== %v 的 maimaiDX 分析报告 ====
底分: %v
b15 天花板: %v
b15 地  板: %v
b25 天花板: %v
b25 地  板: %v
-----
地  板RA: %v，相当于 %v
天花板RA: %v，相当于 %v
`, result.Nickname, result.BaseRa, FormatChartScore(result.DXCeiling), FormatChartScore(result.DXFloor), FormatChartScore(result.SDCeiling), FormatChartScore(result.SDFloor), result.GetFloorRa(), FormatRaSuggestion(result.GetFloorRa()), result.GetCeilingRa(), FormatRaSuggestion(result.GetCeilingRa())))
		return
	})
}
