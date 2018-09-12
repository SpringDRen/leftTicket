package main

import (
	"github.com/SpringDRen/easylog"
	"fmt"
	"os"
)

func main() {
	//输出到文件
	err := easylog.InitFile(easylog.OutputAll, easylog.LevelInfo, "./logs", "example2.log")
	//如果打开文件，最好是在程序退出时关闭
	defer easylog.CloseLog()
	//创建目录或者文件时也许会失败，处理错误
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	easylog.Debug("level debug", "hello world")
	easylog.Info("level info", "hello world")
	easylog.Error("level error", "hello world")
}
