package simplecache

import (
	"bytes"
	"context"
)

type (
	// GetterFunc Get方法没有命中缓存的情况会调用这个方法
	GetterFunc[VT any] func(ctx context.Context, key string) (val VT, err error)

	// MGetterFunc Met方法没有命中缓存的情况会调用这个方法
	MGetterFunc[KT comparable, VT any] func(ctx context.Context, keys []KT, opts ...interface{}) (result map[KT]VT, err error)

	// FmtKeyFunc 使用该方法给格式化key, 例如加前缀, opts 透传MGET 方法的opts参数
	FmtKeyFunc[KT comparable] func(keys []KT, opts ...interface{}) []string

	// SimpleCache 通用缓存方法，KT是key 的类型， VT是值的类型
	SimpleCache[KT comparable, VT any] struct {
		serializer    ISerializer
		operator      ICacheOperator
		Getter        GetterFunc[VT]
		emptyCacheVal []byte
		fmtKeyFn      FmtKeyFunc[KT]
	}
)

const defaultEmptyCacheVal = "-"

func New[KT comparable, VT any]() *SimpleCache[KT, VT] {
	c := &SimpleCache[KT, VT]{
		serializer:    NewJsonSerializer(),
		emptyCacheVal: []byte(defaultEmptyCacheVal),
	}
	return c
}

func (c *SimpleCache[KT, VT]) WithOperator(o ICacheOperator) *SimpleCache[KT, VT] {
	c.operator = o
	return c
}

func (c *SimpleCache[KT, VT]) WithSerializer(s ISerializer) *SimpleCache[KT, VT] {
	c.serializer = s
	return c
}

func (c *SimpleCache[KT, VT]) WithFmtKeyFn(prefixFn FmtKeyFunc[KT]) *SimpleCache[KT, VT] {
	c.fmtKeyFn = prefixFn
	return c
}

func (c *SimpleCache[KT, VT]) WithEmptyVal(val []byte) *SimpleCache[KT, VT] {
	c.emptyCacheVal = val
	return c
}

func (c *SimpleCache[KT, VT]) MGet(ctx context.Context, keys []KT, getter MGetterFunc[KT, VT], opts ...interface{}) (result map[KT]VT, err error) {
	var cacheResult [][]byte
	preKeys := c.fmtKeyFn(keys, opts...)
	if cacheResult, err = c.operator.MGet(preKeys); err != nil {
		return
	}
	result = make(map[KT]VT)
	unHitKeys := make([]KT, 0)
	preUnHitKeys := make([]string, 0)
	for i, bs := range cacheResult {
		if len(bs) == 0 {
			unHitKeys = append(unHitKeys, keys[i])
			preUnHitKeys = append(preUnHitKeys, preKeys[i])
			continue
		}
		if bytes.Compare(bs, c.emptyCacheVal) == 0 {
			continue
		}
		temp := new(VT)
		if err = c.serializer.Unmarshal(bs, temp); err != nil {
			return
		}
		result[keys[i]] = *temp
	}
	if len(unHitKeys) == 0 {
		return
	}

	var getterResult map[KT]VT
	if getterResult, err = c.getterSetCache(ctx, unHitKeys, preUnHitKeys, getter, opts...); err == nil {
		for k, v := range getterResult {
			result[k] = v
		}
	}
	return
}

func (c *SimpleCache[KT, VT]) getterSetCache(ctx context.Context, unHitKeys []KT, preUnHitKeys []string, getter MGetterFunc[KT, VT], opts ...interface{}) (result map[KT]VT, err error) {
	var getterResult map[KT]VT
	if getterResult, err = getter(ctx, unHitKeys, opts...); err != nil {
		return
	}

	result = make(map[KT]VT)
	kvs := make([]interface{}, 0, len(getterResult)*2)
	for key, data := range getterResult {
		var bs []byte
		if bs, err = c.serializer.Marshal(data); err != nil {
			return
		}
		kvs = append(kvs, c.fmtKeyFn([]KT{key}, opts...)[0], bs)
		result[key] = data
	}

	for i, key := range unHitKeys {
		if _, ok := getterResult[key]; ok {
			continue
		}
		kvs = append(kvs, preUnHitKeys[i], c.emptyCacheVal)
	}

	if err = c.operator.MSet(kvs...); err != nil {
		return
	}
	return
}
