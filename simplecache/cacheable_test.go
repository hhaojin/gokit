package simplecache

import (
	"context"
	"fmt"

	"testing"
	"time"
)

func TestSimpleCache_MGet(t *testing.T) {
	var opr ICacheOperator //自定义的缓存操作类，实现 ICacheOperator 接口
	cache := New[int, int](time.Second, &CacheOpts[int]{
		Operator: opr,
	})

	ret, err := cache.MGet(context.Background(), []int{1}, func(ctx context.Context, keys []int) (map[int]int, error) {
		fmt.Println(1)
		return nil, nil
	})
	fmt.Println(ret, err)
}

func TestSimpleCache_Custom(t *testing.T) {

	var opr ICacheOperator //自定义的缓存操作类，实现 ICacheOperator 接口

	preFn := func(keys []int) []string {
		preKeys := make([]string, 0, len(keys))
		for _, key := range keys {
			preKeys = append(preKeys, fmt.Sprintf("pre_%d", key))
		}
		return preKeys
	}
	crossVal := []byte("empty")
	cache := New[int, int](time.Second, &CacheOpts[int]{
		EmptyCacheVal: crossVal,
		FmtKeysFunc:   preFn,
		Serializer:    NewJsonSerializer(),
		Operator:      opr,
		Policy:        NewDefaultPolicy[int](time.Second, opr, crossVal, preFn),
	})

	ret, err := cache.MGet(context.Background(), []int{1}, func(ctx context.Context, keys []int) (map[int]int, error) {
		fmt.Println(1)
		return nil, nil
	})
	fmt.Println(ret, err)
}
