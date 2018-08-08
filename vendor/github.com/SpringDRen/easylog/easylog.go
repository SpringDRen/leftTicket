//Package easylog go 标准包log简单封装.
package easylog

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//输出标志
const (
	OutputConsole = 1 << iota                  //输出到控制台
	OutputFile                                 //输出到文件
	OutputAll     = OutputConsole | OutputFile //同时输出到控制台和文件
)

//日志级别
const (
	LevelDebug = iota
	LevelInfo
	LevelError
)

var (
	outputFlag  int
	console     *log.Logger
	logger      *log.Logger
	currentFile *os.File
	levelFlag   = LevelDebug
	dateFormat  = "2006-01-02 15:04:05.999"
	initFlag    = false
)

//Debug log.
func Debug(v ...interface{}) {
	if LevelDebug >= levelFlag {
		baseOutput("DEBUG", v...)
	}
}

//Info log.
func Info(v ...interface{}) {
	if LevelInfo >= levelFlag {
		baseOutput("INFO", v...)
	}
}

//Error log.
func Error(v ...interface{}) {
	if LevelError >= levelFlag {
		baseOutput("ERROR", v...)
	}
}

func baseOutput(loglevel string, v ...interface{}) {
	if OutputAll&outputFlag != 0 || !initFlag {
		now := time.Now()
		msg := fmt.Sprintln(v...)

		//文件名及行数
		var moduleAndFileName string
		_, fullFilePath, line, _ := runtime.Caller(2)
		dirPath, fileName := path.Split(fullFilePath)
		if dirPath != "" {
			var moduleName string
			dirPath = dirPath[:len(dirPath)-1]
			dirPath, moduleName = path.Split(dirPath)
			moduleAndFileName = moduleName + "/" + fileName
		} else {
			moduleAndFileName = fileName
		}

		timeStr := now.Format(dateFormat)

		if initFlag {
			if OutputConsole&outputFlag != 0 {
				console.Printf("%-23s [%-5s] %s:%d: %s",
					timeStr, loglevel, fileName, line, msg)
			}
			if OutputFile&outputFlag != 0 {
				logger.Printf("%-23s [%-5s] %s:%d: %s",
					timeStr, loglevel, moduleAndFileName, line, msg)
			}
		} else {
			fmt.Printf("%-23s [%-5s] %s:%d: %s",
				timeStr, loglevel, fileName, line, msg)
		}
	}
}

//Init 仅输出到控制台.
func Init(level int) {
	InitFile(OutputConsole, level, "", "")
}

//InitFile 可输出到文件、控制台.
//flag:OutputConsole or OutputFile; level:LevelDebug,LevelInfo,LevelError;
//dir:logfile parent dir;fileName:logfile name
func InitFile(flag int, level int, dir string, fileName string) {
	if initFlag {
		return
	}
	initFlag = true
	outputFlag = flag
	levelFlag = level
	//std console
	if OutputConsole&outputFlag != 0 {
		console = log.New(os.Stdout, "", 0)
	}
	//是否输出到文件
	if OutputFile&outputFlag == 0 {
		return
	}

	//处理路径及日志文件名前缀
	if strings.TrimSpace(dir) == "" {
		dir = "./"
	}
	dir = filepath.ToSlash(dir)
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	if strings.TrimSpace(fileName) == "" {
		arr := strings.Split(os.Args[0], string(os.PathSeparator))
		fileName = arr[len(arr)-1] + ".log"
	}
	logPath := dir + fileName

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_RDONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't open file %s, %v", logPath, err)
		os.Exit(1)
	}
	logger = log.New(file, "", 0)

	return
}

//CloseLog will close current logfile.
func CloseLog() {
	if currentFile != nil {
		currentFile.Close()
	}
}
