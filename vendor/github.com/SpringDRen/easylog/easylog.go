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
	outputFlag  int                         //输出标志
	console     *log.Logger                 //控制台输出
	logger      *log.Logger                 //文件输出
	currentFile *os.File                    //日志文件
	levelFlag   = LevelDebug                //日志级别
	dateFormat  = "2006-01-02 15:04:05.999" //日志时间前缀格式
	initFlag    = false                     //初始化表示
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
			//未初始化，输出到控制台
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
//flag:OutputConsole or OutputFile;
//level:LevelDebug,LevelInfo,LevelError;
//dir:logfile parent dir, empty then use ./
//fileName:logfile name, empty then use os.Args[0]
//mkdir or create/read file may fail.If init success or already init, return nil, otherwise return error
func InitFile(flag int, level int, dir string, fileName string) error {
	if initFlag {
		return nil
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
		return nil
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
	//create log dir
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		initFlag = false
		return fmt.Errorf("mkdir err, %v \n", err)
	}

	logPath := dir + fileName
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		initFlag = false
		return fmt.Errorf("logfile err, %v \n", err)
	}
	logger = log.New(file, "", 0)

	return nil
}

//CloseLog will close current logfile.
func CloseLog() {
	if currentFile != nil {
		currentFile.Close()
	}
}
