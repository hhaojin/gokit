package simplecache

import (
	"context"
	"fmt"
	"testing"
)

func TestSimpleCache_MGet(t *testing.T) {
	var opr ICacheOperator //自定义的缓存操作类，实现 ICacheOperator 接口
	cache := New[int64, *UserDept]().WithOperator(opr)

	ret, err := cache.MGet(context.Background(), []int{1}, func(ctx context.Context, keys []int, opts ...interface{}) (map[int]int, error) {
		fmt.Println(1)
		return nil, nil
	})
	fmt.Println(ret, err)
}

func TestSimpleCache_Custom(t *testing.T) {

	var opr ICacheOperator //自定义的缓存操作类，实现 ICacheOperator 接口
	var ser ISerializer

	cache := New[int64, *UserDept]().
		WithOperator(opr).
		WithSerializer(ser).
		WithEmptyVal([]byte("--")).
		WithFmtKeyFn(func(keys []int64, opts ...interface{}) []string {
			fmt.Println("FmtKeyFn", opts[0], opts[1], opts[2]) // 1 2 3
			
			fmtKeys := make([]string, 0, len(keys))
			for _, key := range keys {
				fmtKeys = append(fmtKeys, fmt.Sprintf("prefix:%v:%d", opts[0], key))
			}
			return fmtKeys
		})

	ret, err := cache.MGet(context.Background(), []int{1}, func(ctx context.Context, keys []int, opts ...interface{}) (map[int]int, error) {
		fmt.Println("MGet", opts[0], opts[1], opts[2]) // 1 2 3
		return nil, nil
	}, 1, 2, 3)
	fmt.Println(ret, err)
}
