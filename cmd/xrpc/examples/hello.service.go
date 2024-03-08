// Code generated by xrpc. DO NOT EDIT.
package main

import "github.com/yc90s/xrpc"


type IHelloService interface {
    Hello(string) (string, error)
    Add(*string, int) (*string, error)
}

func RegisterHelloServiceServer(rpc *xrpc.RPCServer, s IHelloService) {
	rpc.RegisterGO("Hello", s.Hello)
	rpc.Register("Add", s.Add)
}

type HelloServiceClient struct {
    c xrpc.IRPCClient
}

func NewHelloServiceClient(c xrpc.IRPCClient) *HelloServiceClient {
    return &HelloServiceClient{c: c}
}

func (c *HelloServiceClient) Close() {
	c.c.Close()
}

func (c *HelloServiceClient) Hello(subj string, arg0 string) (string, error) {
    var reply string
    err := c.c.Call(subj, "Hello", &reply, arg0)
    return reply, err
}
func (c *HelloServiceClient) Add(subj string, arg0 *string, arg1 int) (*string, error) {
    var reply string
    err := c.c.Call(subj, "Add", &reply, arg0, arg1)
    return &reply, err
}
type IWorldService interface {
    Hi()
    Sum() (int, error)
}

func RegisterWorldServiceServer(rpc *xrpc.RPCServer, s IWorldService) {
	rpc.Register("Hi", s.Hi)
	rpc.RegisterGO("Sum", s.Sum)
}

type WorldServiceClient struct {
    c xrpc.IRPCClient
}

func NewWorldServiceClient(c xrpc.IRPCClient) *WorldServiceClient {
    return &WorldServiceClient{c: c}
}

func (c *WorldServiceClient) Close() {
	c.c.Close()
}

func (c *WorldServiceClient) Hi(subj string) error {
    err := c.c.Cast(subj, "Hi")
    return err
}
func (c *WorldServiceClient) Sum(subj string) (int, error) {
    var reply int
    err := c.c.Call(subj, "Sum", &reply)
    return reply, err
}