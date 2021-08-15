# Example HTTP_Server

## 方式1
```
go build -o http_server main.go
./http_server
```
主进程 和 工作进程 为同一个程序，但是是两个独立的进程。
```
work            52958   1.0  0.1  4989744   4636 s001  S+    1:50下午   0:00.02 ./http_server
work            52889   0.1  0.1  4980924   4692 s001  S+    1:49下午   0:00.02 ./http_server
```

## 方式2
```
go build  -o cmds/httpserver/http_server main.go
gracemaster -conf conf/grace.toml
```

主进程 和 工作进程 为不同程序，是两个独立的进程。
```
worker            53195   0.2  0.1  4988176   4288 s001  S+    1:53下午   0:00.03 gracemaster
worker            53197   0.0  0.1  4982204   4876 s001  S+    1:53下午   0:00.01 ./http_server
```