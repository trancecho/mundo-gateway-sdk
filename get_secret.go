package sdk

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type MyRedisTokenGetter struct {
	rdb *redis.Client
	ctx context.Context
	key string
}

func NewMyRedisTokenGetter(redisAddr, pwd string) *MyRedisTokenGetter {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: pwd,
		DB:       0,
	})
	return &MyRedisTokenGetter{
		rdb: rdb,
		ctx: context.Background(),
		key: "gateway:register:password",
	}
}

func (this *MyRedisTokenGetter) GetToken() (string, error) {
	return this.rdb.Get(this.ctx, this.key).Result()
}
