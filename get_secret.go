package sdk

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type MyRedisTokenGetter struct {
	rdb *redis.Client
	ctx context.Context
	key string
}

func NewMyRedisTokenGetter(redisAddr, pwd string, db int) *MyRedisTokenGetter {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: pwd,
		DB:       db,
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
