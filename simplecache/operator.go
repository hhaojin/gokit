package simplecache

import "time"

type ICacheOperator interface {
	MGet([]string) ([][]byte, error)
	MSet(kvs ...interface{}) error
}

type RedisOpr struct{}

func NewRedisOpr() *RedisOpr {
	return &RedisOpr{}
}

func (s *RedisOpr) MGet(keys []string) ([][]byte, error) {
	ret := make([][]byte, 0)

	return ret, nil
}

func (s *RedisOpr) Set(key string, val []byte, ex time.Duration) error { return nil }

func (s *RedisOpr) MSet(kvs ...interface{}) error {
	return nil
}
