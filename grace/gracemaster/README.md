# GraceMaster

see [examples](../examples/)

## 1 配置文件
```toml
StatusDir = "./var/"

# 保持子进程一直存活，若子进程退出，会自动拉起
Keep = true

# 子进程优雅退出的超时时间，可选配置，默认为 10s
StopTimeout = "10s"

# 新进程启动后，老进程退出前的等待时间，默认为 3s
# StartWait="3s"

# 只工作进程
[Workers.default]
# HomeDir 当前工作进程的工作目录,可选
HomeDir="cmds/http_server/"
# EnvFile 给子进程设置环境变量的文件，可选
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

# 新进程启动后，老进程退出前的等待时间
# 当不配置的时候，将使用全局的配置
# StartWait="3s"

[Workers.sleep]
RootDir="cmds/"
Cmd = "./sleep.sh"
```