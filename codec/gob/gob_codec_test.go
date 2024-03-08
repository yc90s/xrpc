package gobcodec

import (
	"reflect"
	"testing"
)

func TestCodec(t *testing.T) {
	temp_data := "world"
	data1 := []byte("hello")
	testData := []interface{}{
		"hello",
		123,
		temp_data,
		&temp_data,
		data1,
	}
	codec := NewCodec()

	for _, arg := range testData {
		data, err := codec.Marshal(arg)
		if err != nil {
			t.Error(err)
			return
		}

		var rv reflect.Value
		rt := reflect.TypeOf(arg)
		if rt.Kind() == reflect.Ptr {
			rv = reflect.New(rt.Elem())
		} else {
			rv = reflect.New(rt)
		}

		err = codec.Unmarshal(data, rv.Interface())
		if err != nil {
			t.Error(err)
			return
		}

		var reply reflect.Value
		if rt.Kind() == reflect.Ptr {
			reply = rv
		} else {
			reply = rv.Elem()
		}

		if !reflect.DeepEqual(reply.Interface(), arg) {
			t.Errorf("reply:%v != arg:%v", reply, arg)
			return
		}
	}
}
