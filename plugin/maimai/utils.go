package maimai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"

	"github.com/sirupsen/logrus"
)

type MaimaiAnalysisResult struct {
	Nickname  string
	BaseRa    float64
	SDCeiling map[string]interface{}
	SDFloor   map[string]interface{}
	DXCeiling map[string]interface{}
	DXFloor   map[string]interface{}
}

func (r *MaimaiAnalysisResult) GetFloorRa() float64 {
	sdra := r.SDFloor["ra"].(float64)
	dxra := r.DXFloor["ra"].(float64)
	if sdra < dxra {
		return sdra
	}
	return dxra
}

func (r *MaimaiAnalysisResult) GetCeilingRa() float64 {
	sdra := r.SDCeiling["ra"].(float64)
	dxra := r.DXCeiling["ra"].(float64)
	if sdra > dxra {
		return sdra
	}
	return dxra
}

func GetWeightForAchievement(acc float64) float64 {
	var weight float64
	segments := [][]float64{{50, 0}, {60, 5}, {70, 6}, {75, 7}, {80, 7.5}, {90, 8}, {94, 9}, {97, 10.5}, {98, 12.5}, {99, 12.75}, {99.5, 13}, {100, 13.25}, {100.5, 13.5}, {101.1, 14}}
	for _, v := range segments {
		if acc < v[0] {
			weight = v[1]
			break
		}
	}
	return weight
}

func FindDifficultyForRaAcc(ra, acc float64) float64 {
	diff := ra / GetWeightForAchievement(acc) / acc * 100
	diff = math.Ceil(diff*10) / 10
	return diff
}

func CalcRa(difficulty, acc float64) float64 {
	weight := GetWeightForAchievement(acc)
	return math.Floor(weight * difficulty * math.Min(acc, 100.5) / 100)
}

func FormatChartScore(m map[string]interface{}) string {
	return fmt.Sprintf("[%v] %v (%v%% 单曲ra:%v)", m["level"], m["title"], m["achievements"], m["ra"])
}
func FormatRaSuggestion(ra float64) string {
	str := fmt.Sprintf("%v SS+, %v SSS, %v SSS+", FindDifficultyForRaAcc(ra, 99.5), FindDifficultyForRaAcc(ra, 100), FindDifficultyForRaAcc(ra, 100.5))
	if FindDifficultyForRaAcc(ra, 100) >= 14 {
		dict := []string{"（您您您您您？", "（我超！您还是人吗？？？", "（娇娇！！！！", "（我太崇拜你啦"}
		str += dict[rand.Intn(len(dict))]
	}
	return str
}

func GetMinMaxChart(username, qq string) *MaimaiAnalysisResult {
	reqmap := make(map[string]string)
	if len(username) != 0 {
		reqmap["username"] = username
	}
	if len(qq) != 0 {
		reqmap["qq"] = qq
	}

	body, _ := json.Marshal(reqmap)
	cli := &http.Client{}
	resp, err := cli.Post("https://www.diving-fish.com/api/maimaidxprober/query/player", "application/json", bytes.NewReader(body))
	if err != nil {
		logrus.Error("[maimai] query player score: post data failed, ", err)
		return nil
	}
	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error("[maimai] query player score: read data failed, ", err)
		return nil
	}
	resmap := make(map[string]interface{})
	err = json.Unmarshal(res, &resmap)
	result := &MaimaiAnalysisResult{}

	nick, ok := resmap["nickname"].(string)
	if !ok {
		return nil
	}
	result.Nickname = nick
	result.BaseRa = resmap["rating"].(float64)
	dxcharts, _ := resmap["charts"].(map[string]interface{})["dx"]
	sdcharts, _ := resmap["charts"].(map[string]interface{})["sd"]

	// sd-min sd-max dx-min dx-max
	ra_rec := []float64{100000, 0, 100000, 0}
	for _, chart := range sdcharts.([]interface{}) {
		chartmap, _ := chart.(map[string]interface{})
		ra := chartmap["ra"].(float64)
		if ra <= ra_rec[0] {
			ra_rec[0] = ra
			result.SDFloor = chartmap
		}
		if ra > ra_rec[1] {
			ra_rec[1] = ra
			result.SDCeiling = chartmap
		}
	}
	for _, chart := range dxcharts.([]interface{}) {
		chartmap, _ := chart.(map[string]interface{})
		ra := chartmap["ra"].(float64)
		if ra <= ra_rec[2] {
			ra_rec[2] = ra
			result.DXFloor = chartmap
		}
		if ra > ra_rec[3] {
			ra_rec[3] = ra
			result.DXCeiling = chartmap
		}
	}
	return result
}
