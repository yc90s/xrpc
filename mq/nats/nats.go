package natsmq

import (
	"errors"
	"sync"
	"time"

	"github.com/yc90s/xrpc/mq"

	"github.com/nats-io/nats.go"
)

type MQueen struct {
	sub         *nats.Subscription
	conn        *nats.Conn
	wg          sync.WaitGroup
	isSubscribe bool
}

func NewMQueen(conn *nats.Conn) *MQueen {
	return &MQueen{
		conn: conn,
	}
}

func (mq *MQueen) Init() error {
	mq.isSubscribe = false
	return nil
}

func (mq *MQueen) GenerateSubj() string {
	return nats.NewInbox()
}

func (mq *MQueen) Publish(subj string, data []byte) error {
	return mq.conn.Publish(subj, data)
}

func (mq *MQueen) Subscribe(subj string, cb mq.MQCallback) error {
	if mq.isSubscribe {
		return errors.New("already subscribed")
	}

	var err error
	mq.sub, err = mq.conn.SubscribeSync(subj)
	if err != nil {
		return err
	}

	mq.isSubscribe = true

	go func() {
		mq.on_request_handle(subj, cb)
	}()
	return nil
}

func (mq *MQueen) UnSubscribe() error {
	if !mq.isSubscribe {
		return errors.New("not subscribed")
	}
	mq.isSubscribe = false
	mq.sub.Unsubscribe()
	mq.wg.Wait()

	return nil
}

func (mq *MQueen) on_request_handle(subj string, cb mq.MQCallback) {
	mq.wg.Add(1)
	defer mq.wg.Done()

	for {
		msg, err := mq.sub.NextMsg(time.Minute)
		if err != nil {
			if err == nats.ErrTimeout {
				continue
			}

			if !mq.sub.IsValid() && mq.isSubscribe {
				// re-subscribe
				mq.sub, err = mq.conn.SubscribeSync(subj)
				if err != nil {
					cb.Callback(nil, err)
					return
				}
			}

			if mq.sub.IsValid() {
				continue
			}

			// if sub is not valid, return
			return
		}

		cb.Callback(msg.Data, nil)
	}
}
