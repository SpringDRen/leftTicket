package main

import (
	"time"

	"strings"

	"github.com/SpringDRen/easylog"
	"github.com/SpringDRen/leftTicket/lefttk"
	"os"
)

//日志
func initLog() {
	now := time.Now()
	logcfg := lefttk.GetSectionCfg("conf/conf.ini", "logger")
	output := easylog.OutputFile
	outputstr, ok := logcfg["output"]
	if ok && "ALL" == strings.ToUpper(outputstr) {
		output = easylog.OutputAll
	}
	dir, ok := logcfg["dir"]
	if !ok {
		dir = "./"
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, os.ModeDir)
	}
	name, ok := logcfg["name"]
	if !ok {
		name = "leftticket.log"
	}
	level, ok := logcfg["level"]
	if !ok {
		level = "INFO"
	}
	var levelInt int
	if "ERROR" == strings.ToUpper(level) {
		levelInt = easylog.LevelError
	} else if "DEBUG" == strings.ToUpper(level) {
		levelInt = easylog.LevelDebug
	} else {
		levelInt = easylog.LevelInfo
	}

	if "DEBUG" == strings.ToUpper(level) {
		easylog.Debug("output:", outputstr, output, "dir:", dir, ", name:", name, ", level:", level)
	}

	easylog.InitFile(output, levelInt, dir, name+"."+now.Format("2006-01-02"))
}

func main() {
	initLog()
	defer easylog.CloseLog()
	lefttk.StartMission()
}
