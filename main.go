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
	limiter, err := NewLimiter(client, key, 10, 1)
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 15; i++ {
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
	scriptId    string
}

func NewLimiter(cli *redis.Client, key string, maxRequests, timeWindow int) (*Limiter, error) {
	file, err := fs.ReadFile("limit.lua")
	if err != nil {
		return nil, err
	}
	err = cli.Del(context.Background(), key).Err()
	if err != nil {
		return nil, err
	}
	l := &Limiter{
		cli:         cli,
		key:         key,
		maxRequests: maxRequests,
		timeWindow:  timeWindow,
		script:      string(file),
	}
	l.scriptId, err = cli.ScriptLoad(context.Background(), l.script).Result()
	if err != nil {
		return nil, err
	}
	return l, nil

}

func (l *Limiter) Take() (bool, error) {
	eval := l.cli.EvalSha(context.Background(), l.scriptId,
		[]string{l.key}, l.maxRequests, l.timeWindow)
	i, err := eval.Bool()
	if err != nil {
		return false, err
	}
	return i, nil
}
