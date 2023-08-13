这是一个简单的缓存查询器

一、简单使用

```go
package main

import (
	"context"
	"fmt"
	
	"github.com/hhaojin/gokit/simplecache"
)

func main() {
	var opr simplecache.ICacheOperator //自定义的缓存操作类，实现 ICacheOperator 接口
	cache := simplecache.New[int, int](time.Second, &simplecache.CacheOpts[int]{
		Operator: opr,
	})
	//当缓存没有查到的时候，会把没用命中缓存的key传递给闭包执行
	//并且会把闭包返回的结果加入缓存，
	//如果闭包也没有结果返回，会给缓存设置空值，避免缓存穿透
	ret, err := cache.MGet(context.Background(), []int{1}, func(ctx context.Context, keys []int) (map[int]int, error) {
		fmt.Println(1)
		return nil, nil
	})
	fmt.Println(ret, err)
}
```

二、自定义使用

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/hhaojin/gokit/simplecache"
)

func main() {
	var opr simplecache.ICacheOperator //自定义的缓存操作类，实现 ICacheOperator 接口
	preFn := func(keys []int) []string {
		preKeys := make([]string, 0, len(keys))
		for _, key := range keys {
			preKeys = append(preKeys, fmt.Sprintf("pre_%d", key))
		}
		return preKeys
	}
	crossVal := []byte("-")
	policy := simplecache.NewDefaultPolicy[int](time.Second, opr, crossVal, preFn)

	cache := simplecache.New[int, int](time.Second, &simplecache.CacheOpts[int]{
		Operator:      opr,
		EmptyCacheVal: crossVal,                        //缓存穿透时在缓存中设置的空值
		FmtKeysFunc:   preFn,                           //格式化传入的key， 例如加前缀
		Serializer:    simplecache.NewJsonSerializer(), //自定义序列化方式，默认json
		Policy:        policy,                          //自定义策略，查询缓存前，和缓存穿透后调用
	})

	ret, err := cache.MGet(context.Background(), []int{1}, func(ctx context.Context, keys []int) (map[int]int, error) {
		fmt.Println(1)
		return nil, nil
	})
	fmt.Println(ret, err)
}
```
