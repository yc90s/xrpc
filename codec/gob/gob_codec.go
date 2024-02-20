package gobcodec

import (
	"bytes"
	"encoding/gob"
)

type Codec struct {
}

func NewCodec() *Codec {
	return &Codec{}
}

func (c *Codec) Unmarshal(b []byte, dst any) error {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	if err := dec.Decode(dst); err != nil {
		return err
	}

	return nil
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
