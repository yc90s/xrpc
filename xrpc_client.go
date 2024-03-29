package xrpc

import (
	"errors"
	"sync"
	"time"

	gobcodec "github.com/yc90s/xrpc/codec/gob"
	xrpcpb "github.com/yc90s/xrpc/pb"

	"github.com/golang/glog"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

// 用于外部封装的接口
type IRPCClient interface {
	Call(subj string, methodName string, reply any, args ...any) error
	Cast(subj string, methodName string, args ...any) error
}

// RPCClient is a rpc client, it must implement the MQCallback interface
type RPCClient struct {
	opts    Options
	calls   sync.Map
	isValid bool
	mu      sync.Mutex
}

func NewRPCClient(opts ...Option) *RPCClient {
	rpc_client := new(RPCClient)
	rpc_client.calls = sync.Map{}

	// default timeout 3s
	rpc_client.opts.timeout = time.Second * 3

	for _, o := range opts {
		o(&rpc_client.opts)
	}

	if rpc_client.opts.codec == nil {
		rpc_client.opts.codec = gobcodec.NewCodec()
	}

	rpc_client.isValid = true
	err := rpc_client.opts.mq.Subscribe(rpc_client.opts.subj, rpc_client)
	if err != nil {
		rpc_client.isValid = false
	}
	return rpc_client
}

// Call is a method to call a rpc method with reply
// goroutine safe
func (c *RPCClient) Call(subj string, methodName string, reply any, args ...any) error {
	if !c.isValid {
		err := c.retry()
		if err != nil {
			return err
		}
	}
	return c._call(subj, methodName, reply, args...)
}

// Cast is a method to call a rpc method without reply
// goroutine safe
func (c *RPCClient) Cast(subj string, methodName string, args ...any) error {
	if !c.isValid {
		err := c.retry()
		if err != nil {
			return err
		}
	}
	return c._cast(subj, methodName, args...)
}

func (c *RPCClient) _cast(subj string, methodName string, args ...any) error {
	var argsData [][]byte

	for _, arg := range args {
		data, err := c.opts.codec.Marshal(arg)
		if err != nil {
			return err
		}
		argsData = append(argsData, data)
	}
	request := &xrpcpb.Request{
		Method: methodName,
		Params: argsData,
	}

	requestData, err := proto.Marshal(request)
	if err != nil {
		return err
	}

	return c.opts.mq.Publish(subj, requestData)
}

func (c *RPCClient) _call(subj string, methodName string, reply any, args ...any) error {
	var argsData [][]byte

	for _, arg := range args {
		data, err := c.opts.codec.Marshal(arg)
		if err != nil {
			return err
		}
		argsData = append(argsData, data)
	}

	randCid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	cid := randCid.String()
	request := &xrpcpb.Request{
		Cid:     cid,
		ReplyTo: c.opts.subj,
		Method:  methodName,
		Params:  argsData,
	}

	requestData, err := proto.Marshal(request)
	if err != nil {
		return err
	}

	doneChan := make(chan *xrpcpb.Response, 1)

	c.calls.Store(cid, doneChan)
	defer func() {
		c.calls.Delete(cid)
	}()

	err = c.opts.mq.Publish(subj, requestData)
	if err != nil {
		return err
	}

	timeout := time.After(c.opts.timeout)
	select {
	case <-timeout:
		return errors.New("timeout")
	case response := <-doneChan:
		if len(response.Error) > 0 {
			return errors.New(response.Error)
		}
		return c.opts.codec.Unmarshal(response.Result, reply)
	}
}

// must goroutine safe
func (c *RPCClient) Callback(data []byte, mqerr error) {
	if mqerr != nil {
		glog.Error(mqerr)
		// some error happend, should close the client
		c.Close()
		return
	}

	var response xrpcpb.Response
	err := proto.Unmarshal(data, &response)
	if err != nil {
		glog.Error(err)
		return
	}
	if doneChan, ok := c.calls.Load(response.Cid); !ok {
		glog.Error("cid not found: ", response.Cid)
	} else {
		doneChan.(chan *xrpcpb.Response) <- &response
	}
}

func (c *RPCClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isValid {
		c.opts.mq.UnSubscribe()
		c.isValid = false
	}
}

func (c *RPCClient) retry() error {
	if c.isValid {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	// need double check
	if c.isValid {
		return nil
	}

	err := c.opts.mq.Subscribe(c.opts.subj, c)
	if err != nil {
		return err
	}
	glog.Info("retry success", c.opts.subj)
	c.isValid = true
	return nil
}
