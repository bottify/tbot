package maimai

import (
	"fmt"
	"os"
	"tbot/utils"
	"testing"
)

const config = "/srv/tbot_dev/config.yaml"

func init() {
	_, err := os.Stat(config)
	if err != nil {
		fmt.Errorf("stat [%v] failed: %v", config, err)
		panic("must provide config.yaml")
	}
	utils.GetConfig().Init(config)
}

func TestFetchScore(t *testing.T) {
	fmt.Println(GetMinMaxChart("tamce", ""))
	fmt.Println(GetMinMaxChart("", "876472013"))
}

func TestRa(t *testing.T) {
	fmt.Println(CalcRa(12.7, 100.1))
	fmt.Println(CalcRa(13.3, 99.8))
	fmt.Println(FindDifficultyForRaAcc(175, 99.5))
}

func TestMusicData(t *testing.T) {
	data, err := LoadChartData()
	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}
	for _, chart := range data.LevelMap["14+"] {
		t.Logf("%+v\n", *chart)
	}
}

func TestRecommendChart(t *testing.T) {
	chart, diff := GetRecommendChartForRa(180, 99.5)
	t.Logf("expect_ra: %v", CalcRa(diff, 99.5))
	t.Logf("%+v, diffi: %v", *chart, diff)
}
