# easylog

第一个go练手程序。对golang标准包log的简单封装，方便开发测试，不适于用于生产环境。 借鉴于java log4j。

## 支持

1. 指定输出目录，日志文件名，需要先初始化
2. 日志分级输出，Debug,Info,Error，低级别包含高级别
3. 可控制是否输出到控制台、是否输出到文件
4. 固定输出格式： yyyy-MM-dd HH:mm:ss.SSS [level] filename:line: msg
5. 未初始化时，使用标准输出fmt

## 不支持

- 不支持自定义输出串，只能输出 v...
- 非异步输出
- 未充分考虑效率、资源消耗问题
- 无文件切割；不能分文件，不能切换文件
- 不支持配置文件

## 示例

参考examples下的示例.

```go
package main

import (
    "github.com/SpringDRen/easylog"
)

func main() {
    easylog.InitFile(easylog.OutputAll, easylog.LevelInfo,  
      "./", "example2.log")  
    //程序退出时关闭日志文件
    defer easylog.CloseLog()
    easylog.Debug("level debug", "hello world")
    easylog.Info("level info", "hello world")
    easylog.Error("level error", "hello world")
}
```
