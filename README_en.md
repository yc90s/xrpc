# XRPC 
English | [中文](README.md)

[![License](https://img.shields.io/badge/license-apache-blue)](https://opensource.org/licenses/Apache-2.0) 

XRPC is a concise and lightweight RPC framework based on message queues

## Overview
The principle of XRPC design is to implement a lightweight RPC framework based on message queues, which is easy to expand and use. 

Its core is very streamlined, and it provides an interface for message queues and serialization, making it very convenient to customize and expand

## Features
- Use a message queue as the channel for RPC communication.
- Supports remote calls with any number of parameters.
- Supports two remote call methods, `Call` and `Cast`, `Cast` is suitable for situations where no return value needs to be obtained.
- Code generation. Implementing a set of IDL that closely aligns with Go syntax to define interface information for RPC services and automatically generate relevant code.
- Easy to use, with very concise core code.
- Easy to extend, it can easily support various message queues and serialization methods.

## Getting Started
### Install NATS
Install and start NATS, refer to [Install Nats](https://docs.nats.io/running-a-nats-service/introduction/installation)

### Install code generation tool for xrpc
```
go install github.com/yc90s/xrpc/cmd/xrpc@latest
```
After successful installation, execute `xrpc version` to see the specific version number output.

### Define service interfaces
Enter your own project directory and write the following service interface file `hello.service`, which defines the `HelloService` service and includes a `Hello` method.
```
package main 

service HelloService {
    Hello(string) (string, error)
}
```

After running the following command, `xrpc` will generate `hello.service.go` in the current directory based on the content of the `hello.service`, which contains RPC interface information.
```
xrpc hello.service
```

### Code
Next, implement the `HelloService` interface defined above and call the `Hello` method through RPC.
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
- `HelloRPCService` implements the RPC interface we defined, which has a `Hello` method.
- `newHelloService` function is used to create and register an RPC service with `nats`, and subscription to `hello_server`.
- `newHelloRPCServiceClient` function creates an RPC client
- In the `main` function, first start the RPC service through `Start`, then RPC calls the `Hello` method, and finally outputs the result `hello:yc90s`

See the [examples](https://github.com/yc90s/xrpc/tree/master/examples) for more detailed information on usage.

## License
XRPC is Apache 2.0 licensed.
