# enen v0.0.5 -- develop(93c0f8e)
a game framework based on gamelib

## 项目介绍

- 支持热更新、热修复操作
- 支持客户端tcp长连接
- 支持动态扩展
- 支持功能测试、压力测试
- 各服务间采用grpc通信

## 源码编译（需要golang环境）

- make 或者 go build

## 启动命令(参照configs/server.json配置文件)

- center服务：enen center
- gate服务：enen gate
- game服务：enen game
- gate1服务：enen gate -n=gate1
- game1服务：enen game -n=game1

## 测试命令

- 10000个机器人同时在线测试命令：enen test -f=robot -r=user_10000 -d=false

## 命令帮助

- enen -h
- enen [服务] -h
- enen version

## 项目计划

- 由日志转测试脚本工具
- gmt服务

![Image text](https://github.com/laonsx/pngs/blob/master/enen_server_1.png)
