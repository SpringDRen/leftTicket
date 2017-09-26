package lefttk

import (
	"net/smtp"
	"strconv"
	"strings"

	"github.com/SpringDRen/easylog"
)

var (
	//是否发送邮件
	Mailflag = false
	//邮箱配置初始化标志
	initMailFlag = false
	//smtp邮箱host
	host string
	//host：port
	address string
	//发送方邮箱地址
	username string
	//发送方密码
	password string
	//收件人
	to string
	//主题
	subject string
)

//初始化邮件配置，成功返回true,""，失败返回false,msg
func InitMail() (bool, string) {
	if initMailFlag {
		return true, "邮箱配置已经初始化"
	}
	cfg := GetSectionCfg(CONF_FILE, "smtpmail")
	if cfg == nil {
		return false, "读取邮箱配置失败"
	}
	easylog.Debug("mail cfg:", cfg)
	mailbool, ok := cfg["mailflag"]
	if !ok {
		//默认不开启发送邮件
		mailbool = "false"
	}
	var err error
	Mailflag, err = strconv.ParseBool(mailbool)
	if err != nil {
		return false, "mailflag 配置有误"
	}
	if !Mailflag {
		return true, "不发送邮件"
	}
	address, ok = cfg["address"]
	if !ok || address == "" {
		return false, "address 配置有误"
	}
	host = strings.Split(address, ":")[0]
	username, ok = cfg["username"]
	if !ok || username == "" {
		return false, "username 配置有误"
	}
	password, ok = cfg["password"]
	if !ok || password == "" {
		return false, "password 配置有误"
	}
	to, ok = cfg["to"]
	if !ok || to == "" {
		return false, "to 配置有误"
	}

	subject, ok = cfg["subject"]
	if !ok {
		return false, "subject 配置有误"
	}

	initMailFlag = true
	return true, "邮箱配置初始化成功"
}

//发送邮件，如果未初始化，先初始化
func SendText(subjectText, text string) {
	//初始化配置
	if !initMailFlag {
		if flag, msg := InitMail(); !flag {
			easylog.Error(msg)
			return
		} else {
			easylog.Info(msg)
		}
	}
	if !Mailflag {
		easylog.Info("mailflag is false.")
		return
	}
	//认证
	auth := smtp.PlainAuth("", username, password, host)
	//发送
	msg := []byte("To: " + to + "\r\nFrom: " + username + "\r\nSubject: " + subjectText + "-" + subject + "\r\n\r\n" + text)
	err := smtp.SendMail(address, auth, username, []string{to}, msg)
	if err != nil {
		easylog.Error("fail to send mail..", err)
		return
	}
	easylog.Info("mail success.")
}
