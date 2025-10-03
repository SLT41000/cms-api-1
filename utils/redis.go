package utils

import (
	"context"
	"fmt"
	"log"
	"mainPackage/config"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func getExpire() time.Duration {
	expireStr := os.Getenv("CACHE_EXPIRE")
	expireInt, err := strconv.Atoi(expireStr)
	if err != nil {
		expireInt = 10
	}
	expiration := time.Duration(expireInt) * time.Second
	return expiration
}

func RedisSet(key string, value string) error {
	name := os.Getenv("CACHE_PREFIX") + ":" + key
	log.Print(name)
	expiration := getExpire()
	return config.Rdb.Set(context.Background(), name, value, expiration).Err()
}

func RedisGet(key string) (string, error) {
	name := key
	val, err := config.Rdb.Get(context.Background(), name).Result()

	if err == redis.Nil {
		log.Printf("Key '%s' does not exist in Redis\n", name)
		return "", nil // not an error in your logic
	} else if err != nil {
		log.Printf("Redis error getting key '%s': %v\n", name, err)
		return "", err
	}

	log.Printf("Fetched key '%s' = %s\n", name, val)
	return val, nil
}

func RedisDel(key string) error {
	name := os.Getenv("CACHE_PREFIX") + ":" + key
	log.Print("Deleting:", name)
	return config.Rdb.Del(context.Background(), name).Err()
}

// ####====Template=====
func CacheSet(key string, value string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("XXXXX"), key)
	log.Print(name)
	expiration := getExpire()
	return config.Rdb.Set(context.Background(), name, value, expiration).Err()
}

func CacheGet(key string) (string, error) {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("XXXXXX"), key)
	return RedisGet(name)
}

func CacheDel(key string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("XXXXXX"), key)
	log.Print("Deleting:", name)
	return config.Rdb.Del(context.Background(), name).Err()
}

// ####====SLA=====
func OwnerSLASet(value string) error {
	name := fmt.Sprintf("%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_OWNER_SLA"))
	log.Print(name)
	expiration := getExpire()
	return config.Rdb.Set(context.Background(), name, value, expiration).Err()
}

func OwnerSLAGet() (string, error) {
	name := fmt.Sprintf("%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_OWNER_SLA"))
	return RedisGet(name)
}

func OwnerSLADel(key string) error {
	name := fmt.Sprintf("%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_OWNER_SLA"))
	log.Print("Deleting:", name)
	return config.Rdb.Del(context.Background(), name).Err()
}
