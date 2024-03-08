package main

import (
	"errors"
	"flag"

	"github.com/yc90s/xrpc"
	natsmq "github.com/yc90s/xrpc/mq/nats"

	"github.com/golang/glog"
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

func (s *HelloRPCService) HelloError(arg1 *string, arg2 string) (*string, error) {
	return nil, errors.New("hello error")
}

func (s *HelloRPCService) Add(arg1 int, arg2 *int) (int, error) {
	sum := arg1 + *arg2
	return sum, nil
}

func (s *HelloRPCService) Ping() {
	glog.Info("ping")
}

func (s *HelloRPCService) Bye(arg string) {
	glog.Info("bye:" + arg)
}

func newHelloRPCServiceClient(c xrpc.IRPCClient) *HelloServiceClient {
	return NewHelloServiceClient(c)
}

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	nc, err := nats.Connect("nats://127.0.0.1:4222", nats.MaxReconnects(1000))
	if err != nil {
		glog.Error("nats error ", err)
	}
	defer nc.Close()

	s := newHelloService(nc)
	err = s.Start()
	if err != nil {
		glog.Error(err)
	}
	defer s.Stop()

	client := xrpc.NewRPCClient(
		xrpc.SetMQ(natsmq.NewMQueen(nc)),
		xrpc.SetSubj("hello_client"),
	)
	defer client.Close()

	c := newHelloRPCServiceClient(client)

	reply1, err1 := c.Hello("hello_server", "yc90s")
	if err1 != nil {
		glog.Error(err1)
	} else {
		glog.Info(reply1)
	}

	err2 := c.Bye("hello_server", "yc90s")
	if err2 != nil {
		glog.Info(err2)
	}

	err3 := c.Ping("hello_server")
	if err3 != nil {
		glog.Info(err3)
	}

	var arg string = "123"

	// test error
	reply4, err4 := c.HelloError("hello_server", &arg, "456")
	if err4 != nil {
		glog.Error(err4)
	} else {
		glog.Info(*reply4)
	}

	num := 3
	sum, err5 := c.Add("hello_server", 5, &num)
	if err5 != nil {
		glog.Info(err5)
	} else {
		glog.Info(sum)
	}
}
