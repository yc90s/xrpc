package protocodec

import (
	"google.golang.org/protobuf/proto"
)

type Codec struct {
}

func NewCodec() *Codec {
	return &Codec{}
}

func (c *Codec) Unmarshal(b []byte, dst any) error {
	err := proto.Unmarshal(b, dst.(proto.Message))
	if err != nil {
		return err
	}

	return nil
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	b, err := proto.Marshal(v.(proto.Message))
	if err != nil {
		return nil, err
	}

	return b, nil
}
