package xrpc

import (
	"time"

	"github.com/yc90s/xrpc/codec"
	"github.com/yc90s/xrpc/mq"
)

type RequestHeader struct {
	Cid        string
	ReplyTo    string
	MethodName string
}

type Options struct {
	codec   codec.Codec
	mq      mq.MQueen
	subj    string
	timeout time.Duration
}

type Option func(*Options)

func SetCodec(codec codec.Codec) Option {
	return func(o *Options) {
		o.codec = codec
	}
}

func SetMQ(q mq.MQueen) Option {
	return func(o *Options) {
		o.mq = q
	}
}

func SetSubj(subj string) Option {
	return func(o *Options) {
		o.subj = subj
	}
}

func SetTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.timeout = timeout
	}
}
