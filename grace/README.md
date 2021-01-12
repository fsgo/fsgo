# Grace

服务/进程的优雅重启/热加载

## 原理

1. 主进程打开资源句柄（如监听 TCP 端口）
2. 主进程 fork 出子进程，并将打开的资源句柄传给子进程
3. 子进程获取资源句柄，处理资源（如处理 HTTP 请求）
4. 主进程监听信号量 SIGQUIT、SIGUSR2
   1. 收到 SIGUSR2，则
        1. fork 新子进程，处理资源
        2. 老的子进程关闭 ( Graceful )
   2. 收到 SIGQUIT，则
        1. 老的子进程关闭 ( Graceful )
        2. 主进程退出( Start 方法返回)

子进程关闭：
1. 发送 SIGQUIT 给子进程
2. 处理逻辑 Graceful Stop
3. 进程退出

## Example
[examples/http_server/main.go](./examples/http_server/main.go)

1.启动：  
```
./http_server
```
注意：保持这个进程不要退出。

2.reload：  
```
./http_server reload
```
或者：
```
kill -USR2 40332
```
40332 是主进程pid。
