# 简介

通过 webrtc 运行 tcp 服务的一种尝试, 跑起来了, 性能也不是很差

# 测试

```sh
# run server
go run ./cmd/linkport -server
# another terminal
go run ./cmd/linkport -port  :8081
# another terminal
curl http://127.0.0.1:8081
```
