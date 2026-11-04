package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "front-tangelo-forte-41422.db.redis.io:11851",
		Username: "default",
		Password: "OGhrUaeSCTa5soFJoD3jrzeVOb6DTkan",
		DB:       0,
	})

	rdb.Set(ctx, "foo", "bar", 0)
	result, err := rdb.Get(ctx, "foo").Result()

	if err != nil {
		panic(err)
	}

	fmt.Println(result) // >>> bar
}
