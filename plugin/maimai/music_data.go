package maimai

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"tbot/utils"
	"time"

	"github.com/sirupsen/logrus"
)

type ChartInfo struct {
	Difficulty []float64
	Level      []string
	Title      string
	Genre      string
}

func (c *ChartInfo) String() string {
	lvls := strings.Join(c.Level, "/")
	return fmt.Sprintf("[%v][%v]\n%v", lvls, c.Genre, c.Title)
}

type ChartMap struct {
	// level-indexed
	LevelMap      map[string][]*ChartInfo
	DifficultyMap map[float64][]*ChartInfo
}

var mut sync.RWMutex

func GetChartJson() ([]interface{}, error) {
	cached_file := utils.GetConfig().GetDataPath("maimai_data.json")
	var filedata []byte
	info, err := os.Stat(cached_file)
	if err != nil && !os.IsNotExist(err) {
		logrus.Errorf("stat file [%v] failed, %v", cached_file, err)
		return nil, err
	}
	if err != nil || info.ModTime().Add(time.Hour*24*7).Before(time.Now()) {
		mut.Lock()
		defer mut.Unlock()
		logrus.Infof("maimai data expired, updating...")
		cli := &http.Client{}
		resp, err := cli.Get("https://www.diving-fish.com/api/maimaidxprober/music_data")
		if err != nil {
			logrus.Errorf("fetch maimai music data failed: %v", err)
			return nil, err
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.Errorf("fetch maimai music data failed: %v", err)
			return nil, err
		}
		err = os.WriteFile(cached_file, data, os.FileMode(0664))
		if err != nil {
			logrus.Errorf("save maimai music data failed: %v", err)
			return nil, err
		}
		filedata = data
	} else {
		mut.RLock()
		defer mut.RUnlock()
		filedata, err = os.ReadFile(cached_file)
		if err != nil {
			logrus.Errorf("read maimai music data failed: %v", err)
			return nil, err
		}
	}
	parsed := make([]interface{}, 0)
	err = json.Unmarshal(filedata, &parsed)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

var cached_chart_map *ChartMap
var chart_mut sync.Mutex

func LoadChartData() (*ChartMap, error) {
	chart_map := cached_chart_map
	if chart_map != nil {
		return chart_map, nil
	}
	chart_mut.Lock()
	defer chart_mut.Unlock()

	parsed, err := GetChartJson()
	if err != nil {
		return nil, err
	}
	result := &ChartMap{
		LevelMap:      make(map[string][]*ChartInfo),
		DifficultyMap: make(map[float64][]*ChartInfo),
	}
	for _, it := range parsed {
		item, ok := it.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Parse chart info to object failed: %v", it)
		}
		lvl := make([]string, 0)
		for _, lv := range item["level"].([]interface{}) {
			lvl = append(lvl, lv.(string))
		}
		ds := make([]float64, 0)
		for _, d := range item["ds"].([]interface{}) {
			ds = append(ds, d.(float64))
		}
		chart := &ChartInfo{
			Difficulty: ds,
			Level:      lvl,
			Title:      item["title"].(string),
			Genre:      item["basic_info"].(map[string]interface{})["genre"].(string),
		}
		// build index
		for _, d := range ds {
			info, _ := result.DifficultyMap[d]
			if info == nil {
				info = make([]*ChartInfo, 0)
			}
			info = append(info, chart)
			result.DifficultyMap[d] = info
		}
		for _, lv := range lvl {
			info, _ := result.LevelMap[lv]
			if info == nil {
				info = make([]*ChartInfo, 0)
			}
			info = append(info, chart)
			result.LevelMap[lv] = info
		}
	}
	cached_chart_map = result
	return result, nil
}

func GetRecommendChartForRa(ra, acc float64) (*ChartInfo, float64) {
	diffi := FindDifficultyForRaAcc(ra, acc)
	charts, err := LoadChartData()
	if err != nil {
		logrus.Errorf("LoadChartData failed: %v; only recommend difficulty", err)
		return nil, diffi
	}
	if diffi > 15 {
		logrus.Errorf("RecommendChart failed, difficulty too high: %v", diffi)
		return nil, diffi
	}
	for diffi <= 15.0 {
		info, ok := charts.DifficultyMap[diffi]
		if !ok {
			// try higher
			diffi = diffi + 0.1
		}
		return info[rand.Intn(len(info))], diffi
	}
	return nil, diffi
}
