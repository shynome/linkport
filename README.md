# 简介

通过 webrtc 运行 tcp 服务的一种尝试, 跑起来了, 性能也不是很差

ps: 只能在内网中使用, 因为没设置 iceServers

# 测试

```sh
# run server
go run ./cmd/linkport -user test -server -pass vvv -port 127.0.0.1:80
# 不带 -port 运行时会返回内置的http server用以测试连通性
go run ./cmd/linkport -user test -server -pass vvv
# another terminal
go run ./cmd/linkport -user test -port  :8081
# another terminal
curl http://127.0.0.1:8081
```
