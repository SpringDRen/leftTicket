# leftTicket

`leftTicket`是一个基于12306 [web官网](https://kyfw.12306.cn/otn/leftTicket/init) 余票查询程序，并提供两种通知方式：smtp邮件、企业微信号。<br>
基于[queryLeftTicket_end_UAM_js.js?scriptVersion=1.9035](https://kyfw.12306.cn/otn/resources/merged/queryLeftTicket_end_UAM_js.js?scriptVersion=1.9035) 开发。日期：2017-09-26.

- windws 64位下载：[leftTicket.zip](https://github.com/SpringDRen/leftTicket/files/1332251/leftTicket.zip)
- 其他系统自行编译

## Build and run

- sdk：go1.8.3, go1.9 测试均无问题
- dep 本工程使用golang包管理工具dep进行包管理。
- 安装dep `go get -u github.com/golang/dep/cmd/dep`

`go get github.com/SpringDRen/leftTicket` 或者下载压缩包均可（注：解压包之后要将项目放在gopath目录下）。进入此项目目录`cd $GOPATH/github.com/SpringDRen/leftTicket`。

1. 包初始化 `dep ensure`
2. 程序运行依赖配置文件 ./conf/12306.ini、./conf/conf.ini；先修改配置文件
3. 运行或编译 `go run main.go` or `go build`

## 支持

- 定时请求12306，条件过滤余票信息，通知用户
- 使用queryX而不是query进行查询，需要设置header信息；可动态设置header、url等。
- 多任务，可更改任务周期
- 可通过配置更新车站信息
- 可根据乘车时间范围、车次类型、车次、坐席筛选
- smtp邮件通知，可通过配置文件修改
- 微信企业号通知，可通过配置文件修改
- 可更改日志级别、目录等

## 不支持

- 不支持自动抢票，仅是查询
- 需要单机运行；运行完毕需要**手动关闭**
- 不能动态更新配置文件，动态更新任务
- 信息不完全准确，貌似会有缓存；尽管header已设置不使用缓存，但12306后台查询貌似有缓存机制，不能保证信息完全准确，尤其是显示票数为1张的
- 任务没有重试机制，不保证一次查询必然成功，会有404、未返回信息等情况
- 邮件不支持多人发送
- 邮件、微信均不能细分到任务，只能是全局通知
- 日志不能自动切割，不能分日期切割

## 配置文件

详细说明见配置文件注释.

### 12306配置

[12306.ini](./conf/12306.ini)

### 任务、邮箱、微信配置

[conf.ini](./conf/conf.ini)
