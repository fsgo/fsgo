# GraceMaster

see [examples](../examples/)

## 1 配置文件
```toml
StatusDir = "./var/"
# Keep 保持子进程一直存活，若子进程退出，会自动拉起
Keep = true
# 子进程优雅退出的超时时间，单位 ms
StopTimeout = 10000

# 只工作进程
[Workers.default]
# RootDir 当前工作进程的工作目录,可选
RootDir="cmds/http_server/"
# EnvFile给子进程设置环境变量的文件，可选
# 可以是文本文件，需要是 kv 格式，如 "key1=val1"，一行一个
# 也可以是可执行的文件，输出内容为 kv 格式
EnvFile = "prepare.sh"
# Listen 子进程监听的端口，可选
# 若需要热重启功能，则填写，子进程需要使用 grace 的 API 进行开发
Listen = [ "tcp@127.0.0.1:8909", "tcp@127.0.0.1:8910" ]
# Cmd 子进程的启动命令,必填
Cmd = "./http_server"
# CmdArgs 子进程启动命令的参数
CmdArgs=["-conf","../../conf/grace.toml"]

[Workers.sleep]
RootDir="cmds/"
Cmd = "./sleep.sh"
```