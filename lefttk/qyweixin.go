package lefttk

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"bytes"
	"io/ioutil"

	"sync"

	"strconv"

	"github.com/SpringDRen/easylog"
)

var (
	//微信发送标志
	Weixinflag = false
	//初始化标志
	initweixinflag = false
	//企业号id
	corpid string
	//应用 secret
	corpsecret string
	//应用id
	agentid string
	//发送给谁
	touser string
	//缓存的 access_token
	access_token string
	//上一次请求token的时间
	lastTokenTime time.Time
	//token刷新间隔
	tokenDuration = 7200 * time.Second
	locker        sync.Mutex
)

const (
	//access_token url
	CORP_TOKEN_URL = "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=${corpid}&corpsecret=${corpsecret}"
	//send msg url
	CORP_MSG_URL = "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=${access_token}"
)

//初始化微信配置
func InitQyweixin() (bool, string) {
	if initweixinflag {
		return true, "微信配置已经初始化"
	}

	cfg := GetSectionCfg(CONF_FILE, "qyweixin")
	if cfg == nil {
		return false, "读取微信配置失败"
	}
	easylog.Debug("qyweixin cfg:", cfg)
	weixinbool, ok := cfg["weixinflag"]
	if !ok {
		//默认不开启微信
		weixinbool = "false"
	}
	var err error
	Weixinflag, err = strconv.ParseBool(weixinbool)
	if err != nil {
		return false, "weixinflag 配置有误"
	}
	if !Weixinflag {
		return true, "不发送微信"
	}
	corpid, ok = cfg["corpid"]
	if !ok || corpid == "" {
		return false, "corpid 配置有误"
	}
	corpsecret, ok = cfg["corpsecret"]
	if !ok || corpsecret == "" {
		return false, "corpsecret 配置有误"
	}
	agentid, ok = cfg["agentid"]
	if !ok || agentid == "" {
		return false, "agentid 配置有误"
	}
	touser, ok = cfg["touser"]
	if !ok || touser == "" {
		return false, "touser 配置有误"
	}

	initweixinflag = true
	return true, "微信配置初始化成功"
}

//发起一次get access_token，成功返回Access_token，失败返回"",并打印日志
func getAccess_token() string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//处理url
	geturl := strings.Replace(CORP_TOKEN_URL, "${corpid}", corpid, 1)
	geturl = strings.Replace(geturl, "${corpsecret}", corpsecret, 1)
	//构建get请求
	req, err := http.NewRequest(http.MethodGet, geturl, nil)
	if err != nil {
		easylog.Error("new request err..", geturl, err)
		return ""
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		easylog.Error("get url err..", geturl, err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		easylog.Error("can't read respBody..", err)
		return ""
	}
	var rejson AccessTokenMsg
	err = json.Unmarshal(body, &rejson)
	if err == nil {
		if rejson.Errcode == 0 && rejson.Access_token != "" {
			return rejson.Access_token
		} else {
			easylog.Error("errcode:", rejson.Errcode, ", Errmsg:", rejson.Errmsg,
				", access_token:", rejson.Access_token)
		}
	} else {
		easylog.Error("parse access_token json err.", string(body), err)
	}
	return ""
}

//处理access_token，过期则重新获取
func dealAccess_token() {
	locker.Lock()
	defer locker.Unlock()
	//重新获取的标志
	flag := true
	//当access_token不为空是，判断上次获取的时间
	if access_token != "" {
		now := time.Now()
		now = now.Add(-tokenDuration)
		if now.Before(lastTokenTime) {
			flag = false
		} else {
			easylog.Debug("access_token已过期，重新获取")
		}
	}
	//重新请求
	if flag {
		now := time.Now()
		easylog.Debug("开始获取Access_token")
		//最多重试3次
		for i := 0; i < 3; i++ {
			access_token = getAccess_token()
			if access_token != "" {
				break
			}
		}
		lastTokenTime = now
		easylog.Debug("Access_token=", access_token, ", time=", lastTokenTime)
	}
}

//通过企业微信号发送消息
func SendToQyweixin(textmsg string) {
	if !initweixinflag {
		if flag, msg := InitQyweixin(); !flag {
			easylog.Error(msg)
			return
		} else {
			easylog.Info(msg)
		}
	}
	if !Weixinflag {
		easylog.Info("weixinflag is false")
		return
	}
	//处理 access_token
	dealAccess_token()
	if access_token != "" {
		//构建 body json
		bodybytes, err := textMsgBody(textmsg)
		if err != nil {
			easylog.Error("text msg to json err...", err)
			return
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		//处理url
		msgurl := strings.Replace(CORP_MSG_URL, "${access_token}", access_token, 1)
		//构建post请求
		req, err := http.NewRequest(http.MethodPost, msgurl, bytes.NewReader(bodybytes))
		if err != nil {
			easylog.Error("post request err..", msgurl, err)
			return
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Do(req)
		if err != nil {
			easylog.Error("post url err..", msgurl, err)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			easylog.Error("can't read respBody..", err)
			return
		}
		var result ResultMsg
		err = json.Unmarshal(body, &result)
		if err == nil {
			if 0 == result.Errcode {
				easylog.Info("qyweixin send success.")
			} else {
				easylog.Error("qyweixin send fail.", result)
			}
		} else {
			easylog.Error("Unmarshal result json fail.", string(body), err)
		}
	} else {
		easylog.Error("qyweixin send fail.Can't get access_token.")
	}
}

//构建文本信息json串
func textMsgBody(textmsg string) ([]byte, error) {
	var textMsg TextMsg
	textMsg.Touser = touser
	textMsg.Msgtype = "text"
	textMsg.Agentid = agentid
	textMsg.Text.Content = textmsg
	textMsg.Safe = 0
	return json.Marshal(&textMsg)
}

type Text struct {
	Content string `json:"content"`
}

//文本消息
type TextMsg struct {
	//touser	否	成员ID列表（消息接收者，多个接收者用‘|’分隔，最多支持1000个）。特殊情况：指定为@all，则向该企业应用的全部成员发送
	Touser string `json:"touser"`
	//toparty	否	部门ID列表，多个接收者用‘|’分隔，最多支持100个。当touser为@all时忽略本参数
	Toparty string `json:"toparty"`
	//totag	否	标签ID列表，多个接收者用‘|’分隔，最多支持100个。当touser为@all时忽略本参数
	Totag string `json:"totag"`
	//msgtype	是	消息类型，此时固定为：text
	Msgtype string `json:"msgtype"`
	//agentid	是	企业应用的id，整型。可在应用的设置页面查看
	Agentid string `json:"agentid"`
	//text	是	消息内容，最长不超过2048个字节
	Text Text `json:"text"`
	//safe	否	表示是否是保密消息，0表示否，1表示是，默认0
	Safe int `json:"safe"`
}

//access_token 返回消息体
type AccessTokenMsg struct {
	Errcode      int    `json:"errcode"`
	Errmsg       string `json:"errmsg"`
	Access_token string `json:"access_token"`
	Expires_in   int    `json:"expires_in"`
}

//返回消息体
type ResultMsg struct {
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	Invaliduser string `json:"invaliduser"`
}
