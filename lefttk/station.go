package lefttk

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/SpringDRen/easylog"
)

//bjb|北京北|VAP|beijingbei|bjb|0
type Station struct {
	other   string
	chname  string //中文
	code    string //12306需要的code
	pinyin  string
	shortPy string
	index   string
}

const (
	station_nameFile = "station_name.js"
)

var (
	allStations     map[string]Station //所有站点信息，[中文名称]站点信息
	stationsByCode  map[string]Station //[code]站点信息
	initStationFlag = false
)

//https 方式download station_name.js
func downloadStationjs(stationUrl string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(http.MethodGet, stationUrl, nil)
	if err != nil {
		easylog.Error("get station_name.js request err.", err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		easylog.Error("get station_name.js err.", err)
		return
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		easylog.Error("read station err.", err)
		return
	}
	easylog.Debug("get station_name.js success. size=", len(content))
	err = ioutil.WriteFile(station_nameFile, content, 0666)
	if err != nil {
		easylog.Error("can't write station_name.js.", err)
	}
}

//初始化所有站点信息
func initAllStations() {
	easylog.Debug("start init station info")
	//读取配置文件
	cfg := GetSectionCfg(CONF_12306_FILE, "station")
	//更新站点信息标志
	updateStationFlag, ok := cfg["updateStation"]
	if !ok {
		updateStationFlag = "false"
	}

	//文件不存在
	if _, err := os.Stat(station_nameFile); err != nil && os.IsNotExist(err) {
		updateStationFlag = "true"
	}
	//从12306重新download station_name.js
	if "true" == updateStationFlag {
		easylog.Debug("download station_name.js from 12306")
		stationUrl, ok := cfg["url"]
		if !ok {
			easylog.Error(CONF_12306_FILE, "can't find url...")
			return
		}
		downloadStationjs(stationUrl)
	}
	//解析
	allStations = make(map[string]Station)
	stationsByCode = make(map[string]Station)
	jsbytes, err := ioutil.ReadFile(station_nameFile)
	if err != nil {
		easylog.Error("readfile err...", station_nameFile, err)
		return
	}
	tmpstr := string(jsbytes)
	//var station_names ='info';  截取''中的信息
	tmpstr = tmpstr[strings.Index(tmpstr, "'")+1 : strings.LastIndex(tmpstr, "'")]
	//@bjb|北京北|VAP|beijingbei|bjb|0
	for _, v := range strings.Split(tmpstr, "@") {
		if v != "" {
			arr := strings.Split(v, "|")
			var station Station
			station.other = arr[0]
			station.chname = arr[1]
			station.code = arr[2]
			station.pinyin = arr[3]
			station.shortPy = arr[4]
			station.index = arr[5]
			allStations[station.chname] = station
			stationsByCode[station.code] = station
		}
	}
	easylog.Info("finish init station info, allStations size =", len(allStations))
	initStationFlag = true
}

//通过站点中文名称返回code
func GetCodeByChname(chname string) string {
	if !initStationFlag {
		initAllStations()
	}
	if station, ok := allStations[chname]; ok {
		return station.code
	}
	return ""
}

//通过code返回中文名称
func GetChnameByCode(code string) string {
	if !initStationFlag {
		initAllStations()
	}
	if station, ok := stationsByCode[code]; ok {
		return station.chname
	}
	return ""
}
