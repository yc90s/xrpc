// Code generated by xrpc. DO NOT EDIT.
package main

import "github.com/yc90s/xrpc"


type IHelloService interface {
    Hello(string) (string, error)
    HelloError(*string, string) (*string, error)
    Add(int, *int) (int, error)
    Ping()
    Bye(string)
}

func RegisterHelloServiceServer(rpc *xrpc.RPCServer, s IHelloService) {
    rpc.Register("Hello", s.Hello)
    rpc.Register("HelloError", s.HelloError)
    rpc.Register("Add", s.Add)
    rpc.Register("Ping", s.Ping)
    rpc.Register("Bye", s.Bye)
}

type HelloServiceClient struct {
    *xrpc.RPCClient
}

func NewHelloServiceClient(c *xrpc.RPCClient) *HelloServiceClient {
    return &HelloServiceClient{c}
}

func (c *HelloServiceClient) Hello(subj string, arg0 string) (string, error) {
    var reply string
    err := c.Call(subj, "Hello", &reply, arg0)
    return reply, err
}
func (c *HelloServiceClient) HelloError(subj string, arg0 *string, arg1 string) (*string, error) {
    var reply string
    err := c.Call(subj, "HelloError", &reply, arg0, arg1)
    return &reply, err
}
func (c *HelloServiceClient) Add(subj string, arg0 int, arg1 *int) (int, error) {
    var reply int
    err := c.Call(subj, "Add", &reply, arg0, arg1)
    return reply, err
}
func (c *HelloServiceClient) Ping(subj string) error {
    err := c.Cast(subj, "Ping")
    return err
}
func (c *HelloServiceClient) Bye(subj string, arg0 string) error {
    err := c.Cast(subj, "Bye", arg0)
    return err
}