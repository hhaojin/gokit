package simplecache

import "encoding/json"

type ISerializer interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
}

type JsonSerializer struct{}

func NewJsonSerializer() *JsonSerializer {
	return &JsonSerializer{}
}

func (s *JsonSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (s *JsonSerializer) Unmarshal(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}
