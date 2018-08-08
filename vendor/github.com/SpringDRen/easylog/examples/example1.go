package main

import (
	"github.com/SpringDRen/easylog"
)

func main() {
	//仅输出到控制台
	easylog.Init(easylog.LevelDebug)
	easylog.Debug("level debug", "hello world")
	easylog.Info("level info", "hello world")
	easylog.Error("level error", "hello world")
}
