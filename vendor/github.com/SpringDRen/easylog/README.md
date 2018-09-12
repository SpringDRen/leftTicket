# easylog

第一个go练手程序。对golang标准包log的简单封装，方便开发测试，仅go标准包，无任何三方包依赖。借鉴于java log4j。 不适于用于生产环境。

## 支持

1. 指定输出目录、日志文件名，需要先初始化；不存在自动创建目录、文件，失败会返回错误
2. 日志分级输出，Debug,Info,Error，低级别包含高级别
3. 可控制是否输出到控制台、是否输出到文件
4. 固定输出格式： yyyy-MM-dd HH:mm:ss.SSS [level] filename:line: msg
5. 未初始化时，使用标准输出fmt

## 不支持

- 不支持自定义输出串，只能输出 v...
- 非异步输出
- 未充分考虑效率、资源消耗问题
- 无文件切割；不能分文件，不能切换文件
- 不支持配置文件（因为需要添加三方依赖）
- 控制台输出不区分标准、错误，因为idea下会存在乱序问题

## 示例

参考examples下的示例.

```go
package main

import (
    "fmt"
    "os"
    "github.com/SpringDRen/easylog"
)

func main() {
    err := easylog.InitFile(easylog.OutputAll, easylog.LevelInfo,  
      "./", "example2.log")  
    //程序退出时关闭日志文件
    defer easylog.CloseLog()
    //创建目录或者文件时也许会失败，处理错误
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }
    easylog.Debug("level debug", "hello world")
    easylog.Info("level info", "hello world")
    easylog.Error("level error", "hello world")
}
```
