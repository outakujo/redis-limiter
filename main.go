package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/redis/go-redis/v9"
	"sync"
)

//go:embed limit.lua
var fs embed.FS

func main() {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	key := "req_limit"
	// 每秒限制10个
	limiter := NewLimiter(client, key, 10, 1)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			take, _ := limiter.Take()
			fmt.Println(i, take)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

type Limiter struct {
	cli         *redis.Client
	key         string
	maxRequests int
	timeWindow  int
	script      string
}

func NewLimiter(cli *redis.Client, key string, maxRequests, timeWindow int) *Limiter {
	file, err := fs.ReadFile("limit.lua")
	if err != nil {
		panic(err)
	}
	return &Limiter{
		cli:         cli,
		key:         key,
		maxRequests: maxRequests,
		timeWindow:  timeWindow,
		script:      string(file),
	}
}

func (l *Limiter) Take() (bool, error) {
	eval := l.cli.Eval(context.Background(), l.script,
		[]string{l.key}, l.maxRequests, l.timeWindow)
	i, err := eval.Int()
	if err != nil {
		return false, err
	}
	return i == 1, nil
}
