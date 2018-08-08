package main

import (
	"github.com/SpringDRen/easylog"
)

func main() {
	//输出到文件
	easylog.InitFile(easylog.OutputAll, easylog.LevelInfo, "./", "example2.log")
	//如果打开文件，最好是在程序退出时关闭
	defer easylog.CloseLog()
	easylog.Debug("level debug", "hello world")
	easylog.Info("level info", "hello world")
	easylog.Error("level error", "hello world")
}
