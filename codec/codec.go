package codec

type Codec interface {
	Unmarshal(b []byte, dst any) error
	Marshal(v any) ([]byte, error)
}
