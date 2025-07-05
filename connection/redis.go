package connection

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client
var Ctx = context.Background()

func InitRedis() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"), // Contoh: "localhost:6379"
		Password: "",                      // Jika tanpa password
		DB:       0,
	})

	_, err := Redis.Ping(Ctx).Result()
	if err != nil {
		panic("Redis connection failed: " + err.Error())
	}else{
		fmt.Println("💞 Redis terhubung!")
	}

}

func SetToken(key, value string, duration time.Duration) error {
	return Redis.Set(Ctx, key, value, duration).Err()
}

func GetToken(key string) (string, error) {
	return Redis.Get(Ctx, key).Result()
}

func DeleteToken(key string) error {
	return Redis.Del(Ctx, key).Err()
}
