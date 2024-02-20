package xrpc

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	gobcodec "github.com/yc90s/xrpc/codec/gob"
	xrpcpb "github.com/yc90s/xrpc/pb"

	"github.com/golang/glog"
	"google.golang.org/protobuf/proto"
)

var (
	ErrRepeatedRegister  = errors.New("method already registered")
	ErrMethodNotSuitable = errors.New("method not suitable")
)

type MethodInfo struct {
	Method     reflect.Value  // method value
	MethodType reflect.Type   // method type
	InType     []reflect.Type // method args
	OutType    []reflect.Type // method return
	Goroutine  bool
}

type RPCInfo struct {
	request   *xrpcpb.Request
	response  *xrpcpb.Response
	execTime  int64
	needReply bool
}

// RPCServer is a rpc server, it must implement the MQCallback interface
type RPCServer struct {
	opts         Options
	methods      map[string]*MethodInfo
	wg           sync.WaitGroup
	executingNum atomic.Int64 // 正在执行的任务数量
}

func NewRPCServer(opts ...Option) *RPCServer {
	rpc_server := new(RPCServer)
	rpc_server.methods = make(map[string]*MethodInfo)
	for _, o := range opts {
		o(&rpc_server.opts)
	}

	if rpc_server.opts.codec == nil {
		rpc_server.opts.codec = gobcodec.NewCodec()
	}

	if len(rpc_server.opts.subj) == 0 {
		rpc_server.opts.subj = rpc_server.opts.mq.GenerateSubj()
	}

	return rpc_server
}

func (s *RPCServer) _register(name string, f interface{}, goroutine bool) error {
	if _, ok := s.methods[name]; ok {
		return ErrRepeatedRegister
	}

	method := &MethodInfo{
		Method:     reflect.ValueOf(f),
		MethodType: reflect.TypeOf(f),
		Goroutine:  goroutine,
	}

	if !suitableMethod(method.MethodType) {
		return ErrMethodNotSuitable
	}

	method.InType = make([]reflect.Type, method.MethodType.NumIn())
	for i := 0; i < method.MethodType.NumIn(); i++ {
		method.InType[i] = method.MethodType.In(i)
	}

	method.OutType = make([]reflect.Type, method.MethodType.NumOut())
	for i := 0; i < method.MethodType.NumOut(); i++ {
		method.OutType[i] = method.MethodType.Out(i)
	}

	s.methods[name] = method
	return nil
}

func (s *RPCServer) Register(name string, f interface{}) error {
	return s._register(name, f, false)
}

func (s *RPCServer) RegisterGO(name string, f interface{}) error {
	return s._register(name, f, true)
}

func (s *RPCServer) Start() error {
	err := s.opts.mq.Subscribe(s.opts.subj, s)
	if err != nil {
		return err
	}
	return nil
}

func (s *RPCServer) Stop() {
	// unsubscribe
	s.opts.mq.UnSubscribe()

	// wait for executing tasks
	s.wg.Wait()
}

// Callback is the callback function of handle message or error, it must be goroutine safe
func (s *RPCServer) Callback(data []byte, mqerr error) {
	if mqerr != nil {
		glog.Error(mqerr)
		// some error happened, should stop the server
		s.Stop()
		return
	}

	start := time.Now()

	var request xrpcpb.Request
	err := proto.Unmarshal(data, &request)
	if err != nil {
		glog.Info("proto.Unmarshal error: ", err)
		return
	}

	methodInfo, ok := s.methods[request.Method]
	if !ok {
		glog.Info("method not found: ", request.Method)
		return
	}

	if methodInfo.Goroutine {
		go s._runFunc(start, methodInfo, &request)
	} else {
		s._runFunc(start, methodInfo, &request)
	}
}

func (s *RPCServer) _runFunc(start time.Time, methodInfo *MethodInfo, request *xrpcpb.Request) {
	s.wg.Add(1)
	s.executingNum.Add(1)
	defer func() {
		s.wg.Done()
		s.executingNum.Add(-1)
	}()

	if len(request.Params) != len(methodInfo.InType) {
		glog.Info("args num not match: ", request.Method)
		return
	}

	var args = make([]reflect.Value, len(request.Params))
	for k, param := range request.Params {
		var arg reflect.Value
		rt := methodInfo.InType[k]
		if rt.Kind() == reflect.Ptr {
			arg = reflect.New(rt.Elem())
		} else {
			arg = reflect.New(rt)
		}

		err := s.opts.codec.Unmarshal(param, arg.Interface())
		if err != nil {
			glog.Info("Unmarshal error: ", err)
			return
		}

		if rt.Kind() == reflect.Ptr {
			args[k] = arg
		} else {
			args[k] = arg.Elem()
		}
	}

	out := methodInfo.Method.Call(args)

	if len(out) != len(methodInfo.OutType) {
		glog.Info("reply num not match: ", request.Method)
		return
	}

	response := &xrpcpb.Response{
		Cid: request.Cid,
	}

	var needReply bool
	if len(out) > 0 {
		needReply = true
		switch e := out[1].Interface().(type) {
		case nil:
			response.Error = ""
			b, err := s.opts.codec.Marshal(out[0].Interface())
			if err != nil {
				glog.Info("proto.Marshal error: ", err)
				return
			}
			response.Result = b
		case error:
			response.Error = e.Error()
		}
	} else {
		needReply = false
	}

	rpcInfo := &RPCInfo{
		request:   request,
		response:  response,
		execTime:  time.Since(start).Nanoseconds(),
		needReply: needReply,
	}
	s.sendResponse(rpcInfo)
}

func (s *RPCServer) sendResponse(rpcInfo *RPCInfo) {
	if rpcInfo.request.ReplyTo == "" || !rpcInfo.needReply {
		// if replyTo is empty or dont need reply then return
		return
	}

	data, err := proto.Marshal(rpcInfo.response)
	if err != nil {
		glog.Error("proto.Marshal error: ", err)
		return
	}

	err = s.opts.mq.Publish(rpcInfo.request.ReplyTo, data)
	if err != nil {
		glog.Error("mq.Publish error: ", err)
	}
}
