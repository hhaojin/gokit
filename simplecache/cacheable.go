package simplecache

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type (
	GetterFunc[KT comparable, VT any] func(ctx context.Context, keys []KT) (result map[KT]VT, err error)

	SimpleCache[KT comparable, VT any] struct {
		expire        time.Duration
		emptyCacheVal []byte
		fmtKeysFunc   FmtKeysFunc[KT]
		serializer    ISerializer
		operator      ICacheOperator
		policy        IPolicy[KT]
	}

	CacheOpts[KT comparable] struct {
		Operator      ICacheOperator
		EmptyCacheVal []byte
		FmtKeysFunc   FmtKeysFunc[KT]
		Serializer    ISerializer
		Policy        IPolicy[KT]
	}
)

func New[KT comparable, VT any](expire time.Duration, opts *CacheOpts[KT]) *SimpleCache[KT, VT] {
	if opts.Operator == nil {
		panic("opr not be nil")
	}
	c := &SimpleCache[KT, VT]{
		expire:        expire,
		emptyCacheVal: []byte("-"),
		operator:      opts.Operator,
		serializer:    NewJsonSerializer(),
		policy: NewDefaultPolicy[KT](expire, opts.Operator, []byte("-"), func(keys []KT) []string {
			preKeys := make([]string, 0, len(keys))
			for _, key := range keys {
				preKeys = append(preKeys, fmt.Sprintf("%v", key))
			}
			return preKeys
		}),
	}
	if opts.EmptyCacheVal != nil {
		c.emptyCacheVal = opts.EmptyCacheVal
	}
	if opts.FmtKeysFunc != nil {
		c.fmtKeysFunc = opts.FmtKeysFunc
		c.policy = NewDefaultPolicy[KT](expire, c.operator, c.emptyCacheVal, c.fmtKeysFunc)
	}
	if opts.Policy != nil {
		c.policy = opts.Policy
	}
	if opts.Serializer == nil {
		c.serializer = opts.Serializer
	}
	return c
}

func (c *SimpleCache[KT, VT]) MGet(ctx context.Context, keys []KT, getter GetterFunc[KT, VT]) (result map[KT]VT, err error) {
	var unHitKeys []KT
	result, unHitKeys, err = c.getCache(ctx, keys)
	if err != nil {
		return
	}
	err = c.getterSetCache(ctx, unHitKeys, func(ctx context.Context, unHitKeys []KT) (map[KT]VT, error) {
		getterResult, err := getter(ctx, unHitKeys)
		if err != nil {
			return nil, err
		}
		for kt, vt := range getterResult {
			result[kt] = vt
		}
		return getterResult, nil
	})
	return
}

func (c *SimpleCache[KT, VT]) getCache(ctx context.Context, keys []KT) (map[KT]VT, []KT, error) {
	var (
		cacheResult [][]byte
		err         error
	)
	preKeys, err := c.policy.Pre(keys)
	if err != nil {
		return nil, nil, err
	}
	if cacheResult, err = c.operator.MGet(preKeys); err != nil {
		return nil, nil, err
	}
	result := make(map[KT]VT, len(cacheResult))
	if len(cacheResult) == 0 {
		return result, keys, nil
	}
	unHitKeys := make([]KT, 0)
	for i, bs := range cacheResult {
		if len(bs) == 0 {
			unHitKeys = append(unHitKeys, keys[i])
			continue
		}
		if bytes.Compare(bs, c.emptyCacheVal) == 0 {
			continue
		}
		temp := new(VT)
		if err := c.serializer.Unmarshal(bs, temp); err != nil {
			return nil, nil, err
		}
		result[keys[i]] = *temp
	}
	return result, unHitKeys, nil
}

func (c *SimpleCache[KT, VT]) getterSetCache(ctx context.Context, keys []KT, fn GetterFunc[KT, VT]) (err error) {
	if len(keys) == 0 {
		return
	}
	getterResult, err := fn(ctx, keys)
	if err != nil {
		return
	}
	emptyKeys := make([]string, 0, len(getterResult))
	kvs := make([]interface{}, 0, len(getterResult)*2)
	preKeys, err := c.policy.Pre(keys)
	if err != nil {
		return err
	}
	for i, key := range keys {
		preKey := preKeys[i]
		data, ok := getterResult[key]
		if !ok {
			emptyKeys = append(emptyKeys, preKey)
			continue
		}
		var bs []byte
		if bs, err = c.serializer.Marshal(data); err != nil {
			return
		}
		kvs = append(kvs, preKey, bs)
	}
	if err = c.policy.OnCross(emptyKeys); err != nil {
		return err
	}
	if err = c.operator.MSet(kvs...); err != nil {
		return err
	}
	return
}
