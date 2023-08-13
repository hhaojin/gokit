package simplecache

import (
	"time"
)

type FmtKeysFunc[KT comparable] func([]KT) []string

type IPolicy[KT comparable] interface {
	Pre(keys []KT) ([]string, error)
	OnCross(key []string) error
}

type DefaultPolicy[KT comparable] struct {
	expire      time.Duration
	opr         ICacheOperator
	fmtKeysFunc FmtKeysFunc[KT]
	crossVal    []byte
}

type PolicyFn[KT comparable] func(*DefaultPolicy[KT])

func NewDefaultPolicy[KT comparable](expire time.Duration, opr ICacheOperator, crossVal []byte, fn FmtKeysFunc[KT]) *DefaultPolicy[KT] {
	p := &DefaultPolicy[KT]{
		expire:      expire,
		opr:         opr,
		crossVal:    crossVal,
		fmtKeysFunc: fn,
	}
	return p
}

func (p *DefaultPolicy[KT]) OnCross(keys []string) error {
	kvs := make([]interface{}, 0, len(keys)*2)
	for _, key := range keys {
		kvs = append(kvs, key, p.crossVal)
	}
	return p.opr.MSet(kvs...)
}

func (p *DefaultPolicy[KT]) Pre(keys []KT) ([]string, error) {
	return p.fmtKeysFunc(keys), nil
}
