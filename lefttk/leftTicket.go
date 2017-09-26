package lefttk

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"time"

	"strconv"

	"regexp"

	"bytes"

	"github.com/SpringDRen/easylog"
)

const (
	DAY_FORMAT      = "2006-01-02"
	CONF_12306_FILE = "conf/12306.ini"
	CONF_FILE       = "conf/conf.ini"
)

var (
	initConfFlag = false
	queryStart   = "06:00"         // 12306服务时间
	queryEnd     = "23:00"         // 12306服务时间
	queryXurl    string            // 查询票务信息url
	headers      map[string]string //需要的header
	allMissions  []QueryInfo       //所有任务信息
	period       = 2               //查询周期，单位分钟
)

//查询票务信息
func queryX(queryInfo QueryInfo) []byte {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//处理url
	geturl := strings.Replace(queryXurl, "${day}", queryInfo.Day, 1)
	geturl = strings.Replace(geturl, "${from}", queryInfo.FromCode, 1)
	geturl = strings.Replace(geturl, "${to}", queryInfo.ToCode, 1)
	//构建get请求
	req, err := http.NewRequest("GET", geturl, nil)
	if err != nil {
		easylog.Error("queryx get request err.", err)
		return nil
	}
	client := &http.Client{Transport: tr}
	//设置头信息，防止12306校验http来源
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		easylog.Error("queryx get url err..", queryXurl, err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		easylog.Error("queryx can't read respBody..", err)
		return nil
	}
	if body == nil || len(body) < 1 {
		easylog.Error("empty resp.Body. resp.status=", resp.Status)
		return nil
	}
	return body
}

//解析返回的json串
func parseTicketStr(tickets []byte) []TicketInfo {
	//解析json串
	var jsonMap map[string]interface{}
	err := json.Unmarshal(tickets, &jsonMap)
	if err != nil {
		easylog.Error("can't parse json", string(tickets), err)
		return nil
	}
	//取data字段继续解析
	if dataV, ok := jsonMap["data"]; ok {
		switch dataV.(type) {
		case map[string]interface{}:
			//取result继续解析
			result := dataV.(map[string]interface{})["result"]
			switch result.(type) {
			//result 为string数组
			case []interface{}:
				tempArr := result.([]interface{})
				if len(tempArr) > 0 {
					resultArr := make([]TicketInfo, len(tempArr))
					for k, one := range tempArr {
						onestr := one.(string)
						//解析单个字符串，丢弃0,1,取2-35个放入结构体
						arr := strings.Split(onestr, "|")
						var ti TicketInfo
						object := reflect.ValueOf(&ti).Elem()
						for i := 0; i < object.NumField(); i++ {
							object.Field(i).SetString(arr[i+2])
						}
						resultArr[k] = ti
					}
					return resultArr
				} else {
					easylog.Error("result arr len <= 0")
					return nil
				}
			default:
				easylog.Error("result is not []")
				return nil
			}
		default:
			easylog.Error("can't find result. err msg:", jsonMap["messages"])
			return nil
		}
	} else {
		//12306有时候会返回 404
		easylog.Error("没有数据. err message : ", jsonMap["messages"])
		return nil
	}
}

//根据条件过滤，返回条件过滤后的所有票、所有可买列车数
func filterTk(allTks []TicketInfo, queryInfo QueryInfo) (filterTks []TicketInfo, canbuy int) {
	if allTks == nil {
		return nil, 0
	}
	filterTks = make([]TicketInfo, 0, len(allTks))
	for _, tks := range allTks {
		//是否可买
		if strings.ToUpper(tks.CanWebBuy) != "Y" {
			continue
		}
		canbuy++
		//时间
		if queryInfo.Starttime != "" {
			if tks.Start_time < queryInfo.Starttime {
				continue
			}
		}
		if queryInfo.Endtime != "" {
			if tks.Start_time > queryInfo.Endtime {
				continue
			}
		}
		//车次类型
		if queryInfo.Train_type != "" {
			passFlag := false
			for _, trainPrefix := range strings.Split(queryInfo.Train_type, ",") {
				if "QT" == strings.ToUpper(trainPrefix) {
					if ok, _ := regexp.Match("G|C|D|Z|T|K.+", []byte(strings.ToUpper(tks.Station_train_code))); !ok {
						passFlag = true
					}
				} else {
					if strings.HasPrefix(strings.ToUpper(tks.Station_train_code), strings.ToUpper(trainPrefix)) {
						passFlag = true
					}
				}
			}
			if !passFlag {
				continue
			}
		}
		//车次
		if queryInfo.Train != "" {
			passFlag := false
			for _, train := range strings.Split(queryInfo.Train, ",") {
				if strings.ToUpper(train) == strings.ToUpper(tks.Station_train_code) {
					passFlag = true
				}
			}
			if !passFlag {
				continue
			}
		}
		//座位
		if queryInfo.Prior_seat != "" {
			passFlag := false
			for _, seat := range strings.Split(queryInfo.Prior_seat, ",") {
				switch seat {
				case "商务座":
					if tks.Swz_num != "" && tks.Swz_num != "无" {
						passFlag = true
					}
				case "特等座":
					if tks.Tz_num != "" && tks.Tz_num != "无" {
						passFlag = true
					}
				case "一等座":
					if tks.Zy_num != "" && tks.Zy_num != "无" {
						passFlag = true
					}
				case "二等座":
					if tks.Ze_num != "" && tks.Ze_num != "无" {
						passFlag = true
					}
				case "高级软卧":
					if tks.Gr_num != "" && tks.Gr_num != "无" {
						passFlag = true
					}
				case "软卧":
					if tks.Rw_num != "" && tks.Rw_num != "无" {
						passFlag = true
					}
				case "硬卧":
					if tks.Yw_num != "" && tks.Yw_num != "无" {
						passFlag = true
					}
				case "动卧":
					if tks.Srrb_num != "" && tks.Srrb_num != "无" {
						passFlag = true
					}
				case "软座":
					if tks.Rz_num != "" && tks.Rz_num != "无" {
						passFlag = true
					}
				case "硬座":
					if tks.Yz_num != "" && tks.Yz_num != "无" {
						passFlag = true
					}
				case "无座":
					if tks.Wz_num != "" && tks.Wz_num != "无" {
						passFlag = true
					}
				case "其他":
					if tks.Qt_num != "" && tks.Qt_num != "无" {
						passFlag = true
					}
				}
			}
			if !passFlag {
				continue
			}
		} // end 坐席过滤
		filterTks = append(filterTks, tks)

	} //end for range allTks
	return filterTks, canbuy
}

//一次任务查询
func oneJob() {
	//先判断是否在服务时间
	now := time.Now()
	nowStr := now.Format("15:04")
	if nowStr <= queryStart || nowStr >= queryEnd {
		easylog.Debug("非售票时间")
		return
	}
	easylog.Debug("once job start. missions size =", len(allMissions))
	for i, queryOnce := range allMissions {
		easylog.Debug("job", i, "start :", queryOnce)
		ticketstr := queryX(queryOnce)
		if ticketstr == nil {
			continue
		}
		allTickets := parseTicketStr(ticketstr)
		filterTks, canbuyTks := filterTk(allTickets, queryOnce)
		easylog.Info(queryOnce, " job done. all=", len(allTickets),
			", can buy=", canbuyTks, ", filter=", len(filterTks))
		if filterTks != nil && len(filterTks) > 0 {
			//拼接信息
			var textMsgBuf bytes.Buffer
			textMsgBuf.WriteString(queryOnce.Day)
			for _, tk := range filterTks {
				if tkbytes := tk.getTkMsg(queryOnce); tkbytes != nil {
					textMsgBuf.WriteRune('\r')
					textMsgBuf.WriteRune('\n')
					textMsgBuf.Write(tkbytes)
				}
			}
			textMsg := textMsgBuf.String()
			easylog.Info("查询余票成功：", textMsg)
			//发送邮件
			if Mailflag {
				SendText(queryOnce.Day, textMsg)
			}
			//发送微信
			if Weixinflag {
				SendToQyweixin(textMsg)
			}
		}
	} //end for allMissions
	easylog.Info("once job end.")
}

//初始任务配置
func InitConf() (bool, string) {
	if initConfFlag {
		return true, "任务配置已经初始化"
	}
	//##########################12306 相关配置##########################
	//12306 查询 header
	headers = GetSectionCfg(CONF_12306_FILE, "Header")
	easylog.Debug("headers:", headers)
	leftTicketcfg := GetSectionCfg(CONF_12306_FILE, "leftTicket")
	if leftTicketcfg == nil {
		return false, "12306 leftTicket missing"
	}
	easylog.Debug("12306 leftTicket:", leftTicketcfg)
	//网站服务时间
	timestr, ok := leftTicketcfg["time"]
	if !ok {
		timestr = "06:00-23:00"
	}
	timearr := strings.Split(timestr, "-")
	queryStart = timearr[0]
	queryEnd = timearr[1]
	//查询url
	queryXurl, ok = leftTicketcfg["url"]
	if !ok {
		return false, "url 配置有误"
	}
	//##########################任务相关配置##########################
	cfg := GetSectionCfg(CONF_FILE, "leftTicket")
	if cfg == nil {
		return false, "leftTicket section missing"
	}
	easylog.Debug("leftTicket:", cfg)
	//所有任务
	all, ok := cfg["all"]
	if !ok {
		return false, "all 配置有误"
	}
	peristr, ok := cfg["period"]
	if !ok {
		return false, "period 配置有误"
	}
	var err error
	period, err = strconv.Atoi(peristr)
	if err != nil {
		return false, "period 配置有误"
	} else {
		if period < 1 {
			return false, "period 配置有误"
		}
	}
	//初始化所有任务配置
	arr := strings.Split(all, ",")
	allMissions = make([]QueryInfo, len(arr))
	for i, sectionName := range arr {
		onceSecCfg := GetSectionCfg(CONF_FILE, sectionName)
		if onceSecCfg == nil {
			return false, "任务 " + sectionName + " 未配置"
		}
		easylog.Debug(sectionName, " cfg: ", onceSecCfg)
		var queryInfo QueryInfo
		//乘车日期
		if day, ok := onceSecCfg["day"]; !ok {
			return false, sectionName + " day 配置有误"
		} else {
			if _, err := time.Parse(DAY_FORMAT, day); err != nil {
				return false, sectionName + " day 格式不正确，必须为yyyy-MM-dd"
			} else {
				queryInfo.Day = day
			}
		}
		//始发站
		if from, ok := onceSecCfg["from"]; !ok {
			return false, sectionName + " from 配置有误"
		} else {
			queryInfo.From = from
			queryInfo.FromCode = GetCodeByChname(from)
			if queryInfo.FromCode == "" {
				return false, sectionName + " 不存在的车站名：" + from
			}
		}
		//抵达站
		if to, ok := onceSecCfg["to"]; !ok {
			return false, sectionName + " to 配置有误"
		} else {
			queryInfo.To = to
			queryInfo.ToCode = GetCodeByChname(to)
			if queryInfo.ToCode == "" {
				return false, sectionName + " 不存在的车站名：" + to
			}
		}
		queryInfo.Starttime = onceSecCfg["starttime"]
		queryInfo.Endtime = onceSecCfg["endtime"]
		queryInfo.Train_type = onceSecCfg["train_type"]
		queryInfo.Train = onceSecCfg["train"]
		queryInfo.Prior_seat = onceSecCfg["prior_seat"]
		allMissions[i] = queryInfo
	}
	easylog.Debug("allMissions:", allMissions)

	initConfFlag = true
	return true, "任务配置初始化成功"
}

//开始任务
func StartMission() {
	//初始配置
	if ok, msg := InitMail(); !ok {
		easylog.Error("初始邮箱配置失败.", msg)
	} else {
		easylog.Info(msg)
	}
	if ok, msg := InitQyweixin(); !ok {
		easylog.Error("初始微信配置失败.", msg)
	} else {
		easylog.Info(msg)
	}
	if ok, msg := InitConf(); !ok {
		easylog.Error("初始任务配置失败.", msg)
		return
	} else {
		easylog.Info(msg)
	}
	//开启定时任务
	ticker := time.NewTicker(time.Duration(period) * time.Minute)
	easylog.Info("任务开始.周期=", period, "总任务数=", len(allMissions))
	go oneJob()
	for range ticker.C {
		go oneJob()
	}
}

//查询信息
type QueryInfo struct {
	Day        string //日志，yyyy-MM-dd
	From       string //乘车站，中文名
	To         string //抵达站，中文名
	FromCode   string //乘车站code
	ToCode     string //抵达站code
	Starttime  string //发车时间，最小值
	Endtime    string //发车时间，最大值
	Train_type string //火车类型，G-高铁 C-城际 D-动车 Z-直达 T-特快 K-快速 QT-其他
	Train      string //车次
	Prior_seat string //坐席，商务座,特等座,一等座,二等座,高级软卧,软卧,硬卧,动卧,软座,硬座,无座,其他
}

//车次票务信息
type TicketInfo struct {
	Train_no               string //[2]
	Station_train_code     string //[3] 车次号
	Start_station_telecode string //[4] 始发站
	End_station_telecode   string //[5] 终点站
	From_station_telecode  string //[6] 乘车站
	To_station_telecode    string //[7] 抵达车站
	Start_time             string //[8] 乘车时间
	Arrive_time            string //[9] 抵达时间
	Lishi                  string //[10] 耗时
	CanWebBuy              string //[11] Y-可购买，N-不可购买
	Yp_info                string //[12]
	Start_train_date       string //[13]
	Train_seat_feature     string //[14]
	Location_code          string //[15]
	From_station_no        string //[16]
	To_station_no          string //[17]
	Is_support_card        string //[18]
	Controlled_train_flag  string //[19]
	Gg_num                 string //[20]
	Gr_num                 string //[21] 高级软卧
	Qt_num                 string //[22] 其他
	Rw_num                 string //[23] 软卧
	Rz_num                 string //[24] 软座
	Tz_num                 string //[25] 特等座
	Wz_num                 string //[26] 无座
	Yb_num                 string //[27]
	Yw_num                 string //[28] 硬卧
	Yz_num                 string //[29] 硬座
	Ze_num                 string //[30] 二等座
	Zy_num                 string //[31] 一等座
	Swz_num                string //[32] 商务座
	Srrb_num               string //[33] 动卧
	Yp_ex                  string //[34]
	Seat_types             string //[35]
}

//日期 车次 乘车时间乘车站-抵达时间抵达站 坐席信息
func (t *TicketInfo) getTkMsg(queryinfo QueryInfo) []byte {
	var buf bytes.Buffer
	//buf.WriteString(queryinfo.Day)
	//buf.WriteRune(' ')
	buf.WriteString(t.Station_train_code)
	buf.WriteRune(' ')
	buf.WriteString(t.Start_time)
	buf.WriteString(GetChnameByCode(t.From_station_telecode))
	buf.WriteRune('-')
	buf.WriteString(t.Arrive_time)
	buf.WriteString(GetChnameByCode(t.To_station_telecode))
	buf.WriteRune(' ')

	seatFlag := false
	if t.Swz_num != "" && t.Swz_num != "无" {
		if queryinfo.Prior_seat == "" || strings.Contains(queryinfo.Prior_seat, "商务座") {
			buf.WriteString("商务座:")
			buf.WriteString(t.Swz_num)
			buf.WriteRune(',')
			seatFlag = true
		}
	}
	if t.Tz_num != "" && t.Tz_num != "无" {
		if queryinfo.Prior_seat == "" || strings.Contains(queryinfo.Prior_seat, "特等座") {
			buf.WriteString("特等座:")
			buf.WriteString(t.Tz_num)
			buf.WriteRune(',')
			seatFlag = true
		}
	}
	if t.Zy_num != "" && t.Zy_num != "无" {
		if queryinfo.Prior_seat == "" || strings.Contains(queryinfo.Prior_seat, "一等座") {
			buf.WriteString("一等座:")
			buf.WriteString(t.Zy_num)
			buf.WriteRune(',')
			seatFlag = true
		}
	}
	if t.Ze_num != "" && t.Ze_num != "无" {
		if queryinfo.Prior_seat == "" || strings.Contains(queryinfo.Prior_seat, "二等座") {
			buf.WriteString("二等座:")
			buf.WriteString(t.Ze_num)
			buf.WriteRune(',')
			seatFlag = true
		}
	}
	if t.Gr_num != "" && t.Gr_num != "无" {
		if queryinfo.Prior_seat == "" || strings.Contains(queryinfo.Prior_seat, "高级软卧") {
			buf.WriteString("高级软卧:")
			buf.WriteString(t.Gr_num)
			buf.WriteRune(',')
			seatFlag = true
		}
	}
	if t.Rw_num != "" && t.Rw_num != "无" {
		if queryinfo.Prior_seat == "" || strings.Contains(queryinfo.Prior_seat, "软卧") {
			buf.WriteString("软卧:")
			buf.WriteString(t.Rw_num)
			buf.WriteRune(',')
			seatFlag = true
		}
	}
	if t.Yw_num != "" && t.Yw_num != "无" {
		if queryinfo.Prior_seat == "" || strings.Contains(queryinfo.Prior_seat, "硬卧") {
			buf.WriteString("硬卧:")
			buf.WriteString(t.Yw_num)
			buf.WriteRune(',')
			seatFlag = true
		}
	}
	if t.Srrb_num != "" && t.Srrb_num != "无" {
		if queryinfo.Prior_seat == "" || strings.Contains(queryinfo.Prior_seat, "动卧") {
			buf.WriteString("动卧:")
			buf.WriteString(t.Srrb_num)
			buf.WriteRune(',')
			seatFlag = true
		}
	}
	if t.Rz_num != "" && t.Rz_num != "无" {
		if queryinfo.Prior_seat == "" || strings.Contains(queryinfo.Prior_seat, "软座") {
			buf.WriteString("软座:")
			buf.WriteString(t.Rz_num)
			buf.WriteRune(',')
			seatFlag = true
		}
	}
	if t.Yz_num != "" && t.Yz_num != "无" {
		if queryinfo.Prior_seat == "" || strings.Contains(queryinfo.Prior_seat, "硬座") {
			buf.WriteString("硬座:")
			buf.WriteString(t.Yz_num)
			buf.WriteRune(',')
			seatFlag = true
		}
	}
	if t.Wz_num != "" && t.Wz_num != "无" {
		if queryinfo.Prior_seat == "" || strings.Contains(queryinfo.Prior_seat, "无座") {
			buf.WriteString("无座:")
			buf.WriteString(t.Wz_num)
			buf.WriteRune(',')
			seatFlag = true
		}
	}

	if t.Qt_num != "" && t.Qt_num != "无" {
		if queryinfo.Prior_seat == "" || strings.Contains(queryinfo.Prior_seat, "其他") {
			buf.WriteString("其他:")
			buf.WriteString(t.Qt_num)
			buf.WriteRune(',')
			seatFlag = true
		}
	}
	if !seatFlag {
		return nil
	}

	tkbytes := buf.Bytes()
	//去掉多余的,
	if tkbytes[len(tkbytes)-1] == ',' {
		return tkbytes[:len(tkbytes)-1]
	} else {
		return tkbytes
	}
}
