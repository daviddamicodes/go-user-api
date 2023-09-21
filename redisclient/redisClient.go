package redisClient

import (
	"context"

	"github.com/redis/go-redis/v9"
)


func RedisClient() (*redis.Client, error) {
  // DEFAULT CASE FOR LOCALHOST
	client := redis.NewClient(&redis.Options{
    Addr:	  "localhost:6379",
    Password: "", // no password set
    DB:		  0,  // use default DB
  })

  if _, err := client.Ping(context.Background()).Result(); err != nil {
    return nil, err
  }

	return client, nil
}