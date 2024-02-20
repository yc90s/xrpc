package xrpc

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	gobcodec "github.com/yc90s/xrpc/codec/gob"
	xrpcpb "github.com/yc90s/xrpc/pb"

	"github.com/golang/glog"
	"google.golang.org/protobuf/proto"
)

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

	if len(rpc_client.opts.subj) == 0 {
		rpc_client.opts.subj = rpc_client.opts.mq.GenerateSubj()
	}

	rpc_client.isValid = true
	err := rpc_client.opts.mq.Subscribe(rpc_client.opts.subj, rpc_client)
	if err != nil {
		rpc_client.isValid = false
	}
	return rpc_client
}

func generateCID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	cid := hex.EncodeToString(b)
	return cid, nil
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

	cid, err := generateCID()
	if err != nil {
		return err
	}
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
		c.close()
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

func (c *RPCClient) close() {
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
