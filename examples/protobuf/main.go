package main

import (
	"flag"

	"github.com/yc90s/xrpc"
	protocodec "github.com/yc90s/xrpc/codec/proto"
	"github.com/yc90s/xrpc/examples/protobuf/pb"
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
			xrpc.SetCodec(protocodec.NewCodec()),
		),
	}
	RegisterHelloServiceServer(s.RPCServer, s)
	return s
}

func (s *HelloRPCService) Hello(request *pb.String) (*pb.String, error) {
	reply := &pb.String{Value: "hello:" + request.GetValue()}
	return reply, nil
}

func (s *HelloRPCService) Add(arg1 *pb.String, arg2 *pb.String) (*pb.String, error) {
	reply := &pb.String{Value: arg1.GetValue() + arg2.GetValue()}
	return reply, nil
}

func newHelloRPCServiceClient(nc *nats.Conn) *HelloServiceClient {
	return NewHelloServiceClient(xrpc.NewRPCClient(
		xrpc.SetMQ(natsmq.NewMQueen(nc)),
		xrpc.SetSubj("hello_client"),
		xrpc.SetCodec(protocodec.NewCodec()),
	))
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

	c := newHelloRPCServiceClient(nc)
	defer c.Close()

	arg1 := &pb.String{Value: "yc90s"}
	reply1, err1 := c.Hello("hello_server", arg1)
	if err1 != nil {
		glog.Error(err1)
	} else {
		glog.Info(reply1)
	}

	arg2 := &pb.String{Value: "123"}
	reply3, err3 := c.Add("hello_server", arg2, arg2)
	if err3 != nil {
		glog.Info(err3)
	} else {
		glog.Info(reply3)
	}
}
