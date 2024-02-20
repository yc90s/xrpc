# XRPC 
[English](README_en.md) | 中文
[![License](https://img.shields.io/badge/license-apache-blue)](https://opensource.org/licenses/Apache-2.0) 

XRPC是一个基于消息队列的简洁, 轻量的RPC框架

## Overview
XRPC设计的原则是为了实现一套基于消息队列的、易于拓展和易于使用的轻量级RPC框架. 
它的核心非常精简, 而且提供了消息队列的接口, 序列化的接口, 使其可以非常方便地定制化拓展.

## Features
- 使用消息队列作为RPC的通道
- 支持任意参数数量的远程调用
- 支持`Call`和`Cast`两种远程调用方式, `Cast`适用于不需要获取返回值的情况
- 代码生成, 实现了一套IDL, 最大程度贴近go语法, 用来定义rpc服务的接口信息, 并自动生成相关代码
- 容易使用, 核心代码非常精简
- 易拓展, 可以非常容易地支持各种消息队列和各种序列化方式

## Getting Started
### 安装消息队列
安装并启动nats, 参考[安装Nats](https://docs.nats.io/running-a-nats-service/introduction/installation)

### 安装代码生成工具
```
go install github.com/yc90s/xrpc/cmd/xrpc@latest
```
安装成功后执行`xrpc -version`可以看到具体版本号输出

### 定义服务接口
进入自己的项目目录, 然后编写下面的服务接口文件`hello.service`, 它定义了`HelloService`服务, 包含一个`Hello`方法
```
package main 

service HelloService {
    Hello(string) (string, error)
}
```
运行如下命令后, `xrpc`将根据`hello.service`文件内容在当前目录生成`hello.service.go`文件, 里面存放了rpc接口信息
```
xrpc hello.service
```

### 实现代码
接下来实现上面定义的`HelloService`服务接口, 并rpc调用`Hello`方法
```golang
package main

import (
	"fmt"

	"github.com/yc90s/xrpc"
	natsmq "github.com/yc90s/xrpc/mq/nats"

	"github.com/nats-io/nats.go"
)

type HelloRPCService struct {
	*xrpc.RPCServer
}

func newHelloService(nc *nats.Conn) *HelloRPCService {
	s := &HelloRPCService{
		RPCServer: xrpc.NewRPCServer(
			xrpc.SetMQ(natsmq.NewMQueen(nc)),
			xrpc.SetSubj("hello_server"),
		),
	}
	RegisterHelloServiceServer(s.RPCServer, s)
	return s
}

func (s *HelloRPCService) Hello(request string) (string, error) {
	reply := "hello:" + request
	return reply, nil
}

func newHelloRPCServiceClient(nc *nats.Conn) *HelloServiceClient {
	return NewHelloServiceClient(xrpc.NewRPCClient(
		xrpc.SetMQ(natsmq.NewMQueen(nc)),
		xrpc.SetSubj("hello_client"),
	))
}

func main() {
	nc, err := nats.Connect("nats://127.0.0.1:4222", nats.MaxReconnects(1000))
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	s := newHelloService(nc)
	err = s.Start()
	if err != nil {
		panic(err)
	}
	defer s.Stop()

	c := newHelloRPCServiceClient(nc)
	defer c.Close()

	reply, err := c.Hello("hello_server", "yc90s")
	if err != nil {
		panic(err)
	}
	fmt.Println(reply)
}
```
- `HelloRPCService` 实现了我们定义的RPC接口, 它有一个`Hello`方法
- `newHelloService` 函数用来创建并注册RPC服务, 采用`nats`消息队列, 指定订阅`hello_server`主题
- `newHelloRPCServiceClient` 函数创建RPC客户端
- 在`main`函数里面, 首先通过`Start`启动RPC服务, 然后RPC调用`Hello`方法, 最后输出结果`hello:yc90s`

更多的例子可以参考[examples]

## License
XRPC is Apache 2.0 licensed.

## 社区
技术交流群: 384132929