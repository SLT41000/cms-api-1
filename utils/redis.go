package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Rdb *redis.Client
	Ctx = context.Background()
)

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"), // default: no password
		DB:       0,                           // use default DB
	})

	// test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := Rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Println("❌ Redis connection failed:", err)
	} else {
		fmt.Println("✅ Redis connected successfully")
	}
}

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
	return Rdb.Set(context.Background(), name, value, expiration).Err()
}

func RedisGet(key string) (string, error) {
	name := key
	val, err := Rdb.Get(context.Background(), name).Result()

	if err == redis.Nil {
		log.Printf("Key '%s' does not exist in Redis\n", name)
		return "", nil // not an error in your logic
	} else if err != nil {
		log.Printf("Redis error getting key '%s': %v\n", name, err)
		return "", err
	}

	//log.Printf("Fetched key '%s' = %s\n", name, val)
	return val, nil
}

func RedisDel(key string) error {
	name := os.Getenv("CACHE_PREFIX") + ":" + key
	log.Print("Deleting:", name)
	return Rdb.Del(context.Background(), name).Err()
}

// ####====Template=====
func CacheSet(key string, value string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("XXXXX"), key)
	log.Print(name)
	expiration := getExpire()
	return Rdb.Set(context.Background(), name, value, expiration).Err()
}

func CacheGet(key string) (string, error) {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("XXXXXX"), key)
	return RedisGet(name)
}

func CacheDel(key string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("XXXXXX"), key)
	log.Print("Deleting:", name)
	return Rdb.Del(context.Background(), name).Err()
}

// ####====SLA Monitoring=====
func OwnerSLASet(value string) error {
	name := fmt.Sprintf("%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_OWNER_SLA"))
	log.Print(name)
	expiration := getExpire()
	return Rdb.Set(context.Background(), name, value, expiration).Err()
}

func OwnerSLAGet() (string, error) {
	name := fmt.Sprintf("%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_OWNER_SLA"))
	return RedisGet(name)
}

func OwnerSLADel(key string) error {
	name := fmt.Sprintf("%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_OWNER_SLA"))
	log.Print("Deleting:", name)
	return Rdb.Del(context.Background(), name).Err()
}

// ####====Dashboard=====
func GroupTypeSet(value string) error {
	cachePrefix := os.Getenv("CACHE_PREFIX")
	appName := os.Getenv("CACHE_GROUP_TYPE") // ✅ แนะนำให้ใช้ APP_NAME หรือ MODULE_NAME แทน XXXXX
	name := fmt.Sprintf("%s:%s", cachePrefix, appName)

	log.Printf("🔹 CacheSet: %s", name)
	expiration := getExpire()

	return Rdb.Set(context.Background(), name, value, expiration).Err()
}

func GroupTypeGet() (string, error) {
	cachePrefix := os.Getenv("CACHE_PREFIX")
	appName := os.Getenv("CACHE_GROUP_TYPE")
	name := fmt.Sprintf("%s:%s", cachePrefix, appName)

	log.Printf("🔹 CacheGet: %s", name)
	val, err := Rdb.Get(context.Background(), name).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("cache key not found: %s", os.Getenv("CACHE_GROUP_TYPE"))
	}
	return val, err
}

func GroupTypeDet() error {
	cachePrefix := os.Getenv("CACHE_PREFIX")
	appName := os.Getenv("CACHE_GROUP_TYPE")
	name := fmt.Sprintf("%s:%s", cachePrefix, appName)

	log.Printf("🗑️ CacheDel: %s", name)
	return Rdb.Del(context.Background(), name).Err()
}

// ####==== ESB - CREATE WO =====
func EsbCreateSet(key string, value string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_CREATE_WO"), key)
	log.Print(name)
	expiration := getExpire()
	return Rdb.Set(context.Background(), name, value, expiration).Err()
}

func EsbCreateGet(key string) (string, error) {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_CREATE_WO"), key)
	return RedisGet(name)
}

func EsbCreateDel(key string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_CREATE_WO"), key)
	log.Print("Deleting:", name)
	return Rdb.Del(context.Background(), name).Err()
}

// ####==== ESB - UPDATE WO =====
func EsbUpdateSet(key string, value string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_UPDATE_WO"), key)
	log.Print(name)
	//expiration := getExpire()
	return Rdb.Set(context.Background(), name, value, 0).Err()
}

func EsbUpdateGet(key string) (string, error) {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_UPDATE_WO"), key)
	return RedisGet(name)
}

func EsbUpdateDel(key string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_UPDATE_WO"), key)
	log.Print("Deleting:", name)
	return Rdb.Del(context.Background(), name).Err()
}

// ####==== GetUserByUsername =====
func UsernameSet(key string, value string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_USERNAME"), key)
	log.Print(name)
	expiration := getExpire()
	return Rdb.Set(context.Background(), name, value, expiration).Err()
}

func UsernameGet(key string) (string, error) {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_USERNAME"), key)
	return RedisGet(name)
}

func UsernameDel(key string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_USERNAME"), key)
	log.Print("Deleting:", name)
	return Rdb.Del(context.Background(), name).Err()
}

// ####==== GetUserByUsername =====
func UserPermissionSet(key string, value string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_USER_PERMISSION"), key)
	log.Print(name)
	expiration := getExpire()
	return Rdb.Set(context.Background(), name, value, expiration).Err()
}

func UserPermissionGet(key string) (string, error) {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_USER_PERMISSION"), key)
	return RedisGet(name)
}

func UserPermissionDel(key string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_USER_PERMISSION"), key)
	log.Print("Deleting:", name)
	return Rdb.Del(context.Background(), name).Err()
}

// ####====Schedule Monitoring=====
func OwnerScheduleSet(value string) error {
	name := fmt.Sprintf("%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_OWNER_SCHEDULE"))
	log.Print(name)
	expiration := getExpire()
	return Rdb.Set(context.Background(), name, value, expiration).Err()
}

func OwnerScheduleGet() (string, error) {
	name := fmt.Sprintf("%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_OWNER_SCHEDULE"))
	return RedisGet(name)
}

func OwnerScheduleDel() error {
	name := fmt.Sprintf("%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_OWNER_SCHEDULE"))
	log.Print("Deleting:", name)
	return Rdb.Del(context.Background(), name).Err()
}

// ####====Report Monitoring=====
func OwnerReportSet(value string) error {
	name := fmt.Sprintf("%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_OWNER_REPORT"))
	log.Print(name)
	expiration := getExpire()
	return Rdb.Set(context.Background(), name, value, expiration).Err()
}

func OwnerReportGet() (string, error) {
	name := fmt.Sprintf("%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_OWNER_REPORT"))
	return RedisGet(name)
}

func OwnerReportDel(key string) error {
	name := fmt.Sprintf("%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_OWNER_REPORT"))
	log.Print("Deleting:", name)
	return Rdb.Del(context.Background(), name).Err()
}

// ####====Report Monitoring=====
func OwnerDistSet(key, value string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_AREA"), key)
	log.Print(name)
	expiration := getExpire()
	return Rdb.Set(context.Background(), name, value, expiration).Err()
}

func OwnerDistGet(key string) (string, error) {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_AREA"), key)
	return RedisGet(name)
}

func OwnerDistDel(key string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_AREA"), key)
	log.Print("Deleting:", name)
	return Rdb.Del(context.Background(), name).Err()
}

// ####====Stations=====
func OwnerStationSet(key, value string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_STATION"), key)
	log.Print(name)
	expiration := getExpire()
	return Rdb.Set(context.Background(), name, value, expiration).Err()
}

func OwnerStationGet(key string) (string, error) {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_STATION"), key)
	return RedisGet(name)
}

func OwnerStationDel(key string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_STATION"), key)
	log.Print("Deleting:", name)
	return Rdb.Del(context.Background(), name).Err()
}

// ####====Stations=====
func OwnerUserSkillsSet(key, value string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_USER_SKILL"), key)
	log.Print(name)
	expiration := getExpire()
	return Rdb.Set(context.Background(), name, value, expiration).Err()
}

func OwnerUserSkillsGet(key string) (string, error) {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_USER_SKILL"), key)
	return RedisGet(name)
}

func OwnerUserSkillsDel(key string) error {
	name := fmt.Sprintf("%s:%s:%s", os.Getenv("CACHE_PREFIX"), os.Getenv("CACHE_USER_SKILL"), key)
	log.Print("Deleting:", name)
	return Rdb.Del(context.Background(), name).Err()
}
