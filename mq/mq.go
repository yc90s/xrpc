package mq

// MQCallback is the interface that wraps the basic method of a message queue callback.
type MQCallback interface {
	// Callback is the callback function of handle message or error
	Callback([]byte, error)
}

// MQServer is the interface that wraps the basic method of a message queue server.
type MQueen interface {
	// GenerateSubj generates a unique subject name.
	GenerateSubj() string
	// Publish publishes a message to the subject.
	Publish(string, []byte) error
	// Subscribe subscribes a subject.
	Subscribe(string, MQCallback) error
	// UnSubscribe unsubscribes a subject.
	UnSubscribe() error
}
