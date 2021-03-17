/*
A Redis implementation for tinyurl functionality using Redis Labs Managed Service
*/
package redis

import "github.com/go-redis/redis"

func RedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis-12892.c252.ap-southeast-1-1.ec2.cloud.redislabs.com:12892",
		Password: "cK1q1UALUJkEQ0kCSbP4pmgGvmjFQtYk",
		DB:       0,
	})

	return client
}
