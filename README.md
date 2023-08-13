这是一个go工具箱，里面包含了一些日常开发常用的方法，简单易用

一、缓存

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
	//ret: map[int]int
	fmt.Println(ret, err)
	
	//更多使用方法请查看子目录里面的readme文件
}
```
