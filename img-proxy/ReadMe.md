## 介绍
该项目是一个图片代理服务器程序，用于处理源图片数据转发
### 编译
```go
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o img-proxy-server .
```
### 启动
```shell
 ./img-img-proxy-server -port=8081
```