package redisClient

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client
var Ctx = context.Background()

func InitRedis(){
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		log.Fatal("REDIS_URL is not set")
	}

	opt,err := redis.ParseURL(redisUrl)
	if err != nil {
		log.Fatal(err)
	}

	Rdb = redis.NewClient(opt)
	
	_,err = Rdb.Ping(Ctx).Result()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Redis client initialized successfully")
}
